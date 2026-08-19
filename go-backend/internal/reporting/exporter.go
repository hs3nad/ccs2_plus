package reporting

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"ccs2plus/go-backend/internal/store"

	"github.com/xuri/excelize/v2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Uploader interface {
	Upload(ctx context.Context, localPath, remoteName string) (string, error)
}

type DailyExporter struct {
	store          *store.Store
	exportDir      string
	location       *time.Location
	clock          func() time.Time
	staticUploader Uploader
}

type GoogleDriveUploader struct {
	service  *drive.Service
	folderID string
}

func NewDailyExporter(store *store.Store, exportDir string, location *time.Location, uploader Uploader) *DailyExporter {
	return &DailyExporter{
		store:          store,
		exportDir:      exportDir,
		location:       location,
		clock:          time.Now,
		staticUploader: uploader,
	}
}

func NewGoogleDriveUploader(ctx context.Context, credentialsPath, folderID string) (*GoogleDriveUploader, error) {
	svc, err := drive.NewService(
		ctx,
		option.WithCredentialsFile(credentialsPath),
		option.WithScopes(drive.DriveFileScope),
	)
	if err != nil {
		return nil, fmt.Errorf("create google drive service: %w", err)
	}
	return &GoogleDriveUploader{service: svc, folderID: folderID}, nil
}

func (u *GoogleDriveUploader) Upload(ctx context.Context, localPath, remoteName string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open export file: %w", err)
	}
	defer file.Close()

	meta := &drive.File{
		Name:    remoteName,
		Parents: []string{u.folderID},
	}
	created, err := u.service.Files.Create(meta).
		SupportsAllDrives(true).
		Media(file).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("upload to google drive: %w", err)
	}
	return created.Id, nil
}

func (e *DailyExporter) Start(ctx context.Context) {
	if e == nil || e.store == nil {
		return
	}

	go func() {
		e.runPending(ctx)

		for {
			next := nextRunAt(e.clock(), e.location)
			timer := time.NewTimer(time.Until(next))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				e.runPending(ctx)
			}
		}
	}()
}

func (e *DailyExporter) runPending(ctx context.Context) {
	dates, err := e.store.CandidateDailyExportDates(e.clock(), e.location)
	if err != nil {
		log.Printf("daily export candidate lookup failed: %v", err)
	} else {
		for _, day := range dates {
			if err := e.ExportDate(ctx, day); err != nil {
				log.Printf("daily export failed for %s: %v", day.Format("2006-01-02"), err)
			}
		}
	}

	months, err := e.store.CandidateMonthlyExportMonths(e.clock(), e.location)
	if err != nil {
		log.Printf("monthly export candidate lookup failed: %v", err)
		return
	}
	for _, month := range months {
		if err := e.ExportMonth(ctx, month); err != nil {
			log.Printf("monthly export failed for %s: %v", month.Format("2006-01"), err)
		}
	}
}

func (e *DailyExporter) ExportDate(ctx context.Context, day time.Time) error {
	if err := os.MkdirAll(e.exportDir, 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	data, err := e.store.DailyWorkbookData(day, e.location)
	if err != nil {
		_ = e.store.MarkExportRun(day, "failed", "", "", err.Error(), nil)
		return fmt.Errorf("load daily workbook data: %w", err)
	}

	localPath := filepath.Join(e.exportDir, workbookName(day))
	if err := writeWorkbook(localPath, data, e.location); err != nil {
		_ = e.store.MarkExportRun(day, "failed", localPath, "", err.Error(), nil)
		return fmt.Errorf("write workbook: %w", err)
	}

	uploader, ok, err := e.uploader(ctx)
	if err != nil {
		_ = e.store.MarkExportRun(day, "failed", localPath, "", err.Error(), nil)
		return err
	}
	if !ok {
		if err := e.store.MarkExportRun(day, "exported_local", localPath, "", "", nil); err != nil {
			return err
		}
		return nil
	}

	driveFileID, err := uploader.Upload(ctx, localPath, filepath.Base(localPath))
	if err != nil {
		_ = e.store.MarkExportRun(day, "failed", localPath, "", err.Error(), nil)
		return err
	}

	now := e.clock()
	if err := e.store.MarkExportRun(day, "uploaded", localPath, driveFileID, "", &now); err != nil {
		return err
	}
	return nil
}

func (e *DailyExporter) ExportMonth(ctx context.Context, month time.Time) error {
	if err := os.MkdirAll(e.exportDir, 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	data, err := e.store.MonthlyWorkbookData(month, e.location)
	if err != nil {
		_ = e.store.MarkReportRun("monthly", month.Format("2006-01"), "failed", "", "", err.Error(), nil)
		return fmt.Errorf("load monthly workbook data: %w", err)
	}

	localPath := filepath.Join(e.exportDir, monthlyWorkbookName(month))
	if err := writeMonthlyWorkbook(localPath, data, e.location); err != nil {
		_ = e.store.MarkReportRun("monthly", month.Format("2006-01"), "failed", localPath, "", err.Error(), nil)
		return fmt.Errorf("write monthly workbook: %w", err)
	}

	uploader, ok, err := e.uploader(ctx)
	if err != nil {
		_ = e.store.MarkReportRun("monthly", month.Format("2006-01"), "failed", localPath, "", err.Error(), nil)
		return err
	}
	if !ok {
		return e.store.MarkReportRun("monthly", month.Format("2006-01"), "exported_local", localPath, "", "", nil)
	}

	driveFileID, err := uploader.Upload(ctx, localPath, filepath.Base(localPath))
	if err != nil {
		_ = e.store.MarkReportRun("monthly", month.Format("2006-01"), "failed", localPath, "", err.Error(), nil)
		return err
	}

	now := e.clock()
	return e.store.MarkReportRun("monthly", month.Format("2006-01"), "uploaded", localPath, driveFileID, "", &now)
}

func (e *DailyExporter) uploader(ctx context.Context) (Uploader, bool, error) {
	if e.staticUploader != nil {
		return e.staticUploader, true, nil
	}

	cfg, err := e.store.GetGoogleDriveConfig()
	if err != nil {
		return nil, false, err
	}
	if !cfg.Enabled || cfg.CredentialsPath == "" || cfg.FolderID == "" {
		return nil, false, nil
	}
	uploader, err := NewGoogleDriveUploader(ctx, cfg.CredentialsPath, cfg.FolderID)
	if err != nil {
		return nil, false, err
	}
	return uploader, true, nil
}

func nextRunAt(now time.Time, loc *time.Location) time.Time {
	localNow := now.In(loc)
	nextDay := localNow.AddDate(0, 0, 1)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 10, 0, 0, loc)
}

func workbookName(day time.Time) string {
	return fmt.Sprintf("ccs2plus-daily-%s.xlsx", day.Format("2006-01-02"))
}

func monthlyWorkbookName(month time.Time) string {
	return fmt.Sprintf("ccs2plus-monthly-%s.xlsx", month.Format("2006-01"))
}

func writeWorkbook(path string, data store.DailyWorkbookData, loc *time.Location) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	f.SetSheetName("Sheet1", "Summary")
	if err := writeSummarySheet(f, data, loc); err != nil {
		return err
	}
	if err := writePaymentMethodSheet(f, "PaymentByMethod", data.Payments); err != nil {
		return err
	}
	if err := writeTopRoomsSheet(f, "TopRooms", data.Stays); err != nil {
		return err
	}
	if err := writeStaysSheet(f, data, loc); err != nil {
		return err
	}
	if err := writePaymentsSheet(f, data, loc); err != nil {
		return err
	}
	if err := writeAuditSheet(f, data, loc); err != nil {
		return err
	}
	if err := writeRoomsSheet(f, data, loc); err != nil {
		return err
	}
	return f.SaveAs(path)
}

func writeMonthlyWorkbook(path string, data store.MonthlyWorkbookData, loc *time.Location) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	f.SetSheetName("Sheet1", "Summary")
	if err := writeMonthlySummarySheet(f, data, loc); err != nil {
		return err
	}
	if err := writeMonthlyDailySummarySheet(f, data, loc); err != nil {
		return err
	}
	if err := writeMonthlyStayModeSheet(f, data); err != nil {
		return err
	}
	if err := writePaymentMethodSheet(f, "PaymentByMethod", data.Payments); err != nil {
		return err
	}
	if err := writeTopRoomsSheet(f, "TopRooms", data.Stays); err != nil {
		return err
	}
	dailyAsSingle := store.DailyWorkbookData{
		Date:     data.MonthKey,
		Rooms:    data.Rooms,
		Stays:    data.Stays,
		Payments: data.Payments,
		Audits:   data.Audits,
	}
	if err := writeStaysSheet(f, dailyAsSingle, loc); err != nil {
		return err
	}
	if err := writePaymentsSheet(f, dailyAsSingle, loc); err != nil {
		return err
	}
	if err := writeAuditSheet(f, dailyAsSingle, loc); err != nil {
		return err
	}
	if err := writeRoomsSheet(f, dailyAsSingle, loc); err != nil {
		return err
	}
	return f.SaveAs(path)
}

func writeSummarySheet(f *excelize.File, data store.DailyWorkbookData, loc *time.Location) error {
	totalPayment := 0.0
	for _, item := range data.Payments {
		totalPayment += item.Amount
	}
	rows := [][]any{
		{"รายงานการใช้ห้องพักประจำวัน"},
		{"วันที่", data.Date},
		{"เขตเวลา", loc.String()},
		{"จำนวนรายการเข้าพัก", len(data.Stays)},
		{"จำนวนรายการชำระเงิน", len(data.Payments)},
		{"ยอดรับชำระรวม", totalPayment},
		{"จำนวนเหตุการณ์ตรวจสอบ", len(data.Audits)},
		{"จำนวนห้องในระบบ", len(data.Rooms)},
	}
	return writeRows(f, "Summary", rows)
}

func writeMonthlySummarySheet(f *excelize.File, data store.MonthlyWorkbookData, loc *time.Location) error {
	totalPayment := 0.0
	for _, item := range data.Payments {
		totalPayment += item.Amount
	}
	rows := [][]any{
		{"รายงานการใช้ห้องพักประจำเดือน"},
		{"เดือน", data.MonthKey},
		{"เขตเวลา", loc.String()},
		{"จำนวนรายการเข้าพัก", len(data.Stays)},
		{"จำนวนรายการชำระเงิน", len(data.Payments)},
		{"ยอดรับชำระรวม", totalPayment},
		{"จำนวนเหตุการณ์ตรวจสอบ", len(data.Audits)},
		{"จำนวนห้องในระบบ", len(data.Rooms)},
	}
	return writeRows(f, "Summary", rows)
}

func writeMonthlyDailySummarySheet(f *excelize.File, data store.MonthlyWorkbookData, loc *time.Location) error {
	sheet := "DailySummary"
	f.NewSheet(sheet)
	type totals struct {
		stays  int
		pays   int
		amount float64
		audits int
	}
	byDay := map[string]*totals{}
	for _, item := range data.Stays {
		for _, ts := range []*time.Time{item.GuestInAt, item.GuestOutAt, item.CleanUpAt, &item.LastUpdatedAt} {
			if ts == nil {
				continue
			}
			key := ts.In(loc).Format("2006-01-02")
			if byDay[key] == nil {
				byDay[key] = &totals{}
			}
			byDay[key].stays++
			break
		}
	}
	for _, item := range data.Payments {
		key := item.PaidAt.In(loc).Format("2006-01-02")
		if byDay[key] == nil {
			byDay[key] = &totals{}
		}
		byDay[key].pays++
		byDay[key].amount += item.Amount
	}
	for _, item := range data.Audits {
		key := item.Timestamp.In(loc).Format("2006-01-02")
		if byDay[key] == nil {
			byDay[key] = &totals{}
		}
		byDay[key].audits++
	}
	keys := make([]string, 0, len(byDay))
	for key := range byDay {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := [][]any{{"วันที่", "จำนวนเข้าพัก", "จำนวนชำระเงิน", "ยอดรับชำระรวม", "จำนวนเหตุการณ์"}}
	for _, key := range keys {
		item := byDay[key]
		rows = append(rows, []any{key, item.stays, item.pays, item.amount, item.audits})
	}
	return writeRows(f, sheet, rows)
}

func writeMonthlyStayModeSheet(f *excelize.File, data store.MonthlyWorkbookData) error {
	sheet := "StayModeSummary"
	f.NewSheet(sheet)
	type totals struct {
		count  int
		amount float64
	}
	byMode := map[string]*totals{}
	for _, item := range data.Stays {
		mode := item.StayMode
		if mode == "" {
			mode = "unknown"
		}
		if byMode[mode] == nil {
			byMode[mode] = &totals{}
		}
		byMode[mode].count++
		byMode[mode].amount += item.PaymentAmount
	}
	keys := make([]string, 0, len(byMode))
	for key := range byMode {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := [][]any{{"ประเภทการเข้าพัก", "จำนวนเข้าพัก", "ยอดรับชำระรวม"}}
	for _, key := range keys {
		item := byMode[key]
		rows = append(rows, []any{key, item.count, item.amount})
	}
	return writeRows(f, sheet, rows)
}

func writePaymentMethodSheet(f *excelize.File, sheet string, payments []store.PaymentRecord) error {
	f.NewSheet(sheet)
	type totals struct {
		count  int
		amount float64
	}
	byMethod := map[string]*totals{}
	for _, item := range payments {
		method := item.Method
		if method == "" {
			method = "ไม่ระบุ"
		}
		if byMethod[method] == nil {
			byMethod[method] = &totals{}
		}
		byMethod[method].count++
		byMethod[method].amount += item.Amount
	}
	keys := make([]string, 0, len(byMethod))
	for key := range byMethod {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := [][]any{{"วิธีชำระเงิน", "จำนวนรายการ", "ยอดรับชำระรวม"}}
	for _, key := range keys {
		item := byMethod[key]
		rows = append(rows, []any{key, item.count, item.amount})
	}
	return writeRows(f, sheet, rows)
}

func writeTopRoomsSheet(f *excelize.File, sheet string, stays []store.StayRecord) error {
	f.NewSheet(sheet)
	type totals struct {
		count   int
		amount  float64
		minutes int
	}
	byRoom := map[string]*totals{}
	for _, item := range stays {
		roomID := item.RoomID
		if roomID == "" {
			roomID = "ไม่ระบุ"
		}
		if byRoom[roomID] == nil {
			byRoom[roomID] = &totals{}
		}
		byRoom[roomID].count++
		byRoom[roomID].amount += item.PaymentAmount
		byRoom[roomID].minutes += item.UsedMinutes
	}
	keys := make([]string, 0, len(byRoom))
	for key := range byRoom {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := byRoom[keys[i]]
		right := byRoom[keys[j]]
		if left.count == right.count {
			return keys[i] < keys[j]
		}
		return left.count > right.count
	})
	rows := [][]any{{"ลำดับ", "ห้อง", "จำนวนครั้งใช้งาน", "นาทีใช้งานรวม", "ยอดรับชำระรวม"}}
	for idx, key := range keys {
		item := byRoom[key]
		rows = append(rows, []any{idx + 1, key, item.count, item.minutes, item.amount})
	}
	return writeRows(f, sheet, rows)
}

func writeStaysSheet(f *excelize.File, data store.DailyWorkbookData, loc *time.Location) error {
	sheet := "Stays"
	f.NewSheet(sheet)
	rows := [][]any{{
		"รหัสเข้าพัก", "รหัสคำขอ", "ห้อง", "ประเภทการเข้าพัก", "เคส", "สถานะ",
		"เวลาเข้า", "เวลาออก", "เวลาทำความสะอาดเสร็จ", "นาทีที่กำหนด", "นาทีใช้งาน",
		"ยอดรับชำระ", "เลขอ้างอิงชำระเงิน", "หมายเหตุ", "อัปเดตล่าสุด",
	}}
	for _, item := range data.Stays {
		rows = append(rows, []any{
			item.StayID,
			item.RequestID,
			item.RoomID,
			item.StayMode,
			item.CaseCode,
			item.Status,
			formatTime(item.GuestInAt, loc),
			formatTime(item.GuestOutAt, loc),
			formatTime(item.CleanUpAt, loc),
			item.AllocatedMinutes,
			item.UsedMinutes,
			item.PaymentAmount,
			item.PaymentRef,
			item.Remark,
			item.LastUpdatedAt.In(loc).Format(time.RFC3339),
		})
	}
	return writeRows(f, sheet, rows)
}

func writePaymentsSheet(f *excelize.File, data store.DailyWorkbookData, loc *time.Location) error {
	sheet := "Payments"
	f.NewSheet(sheet)
	rows := [][]any{{"รหัสชำระเงิน", "รหัสคำขอ", "รหัสเข้าพัก", "ห้อง", "ยอดชำระ", "วิธีชำระเงิน", "เลขอ้างอิง", "เวลาชำระ"}}
	for _, item := range data.Payments {
		rows = append(rows, []any{
			item.PaymentID,
			item.RequestID,
			item.StayID,
			item.RoomID,
			item.Amount,
			item.Method,
			item.Reference,
			item.PaidAt.In(loc).Format(time.RFC3339),
		})
	}
	return writeRows(f, sheet, rows)
}

func writeAuditSheet(f *excelize.File, data store.DailyWorkbookData, loc *time.Location) error {
	sheet := "Audit"
	f.NewSheet(sheet)
	rows := [][]any{{"รหัสเหตุการณ์", "เวลา", "ห้อง", "การทำรายการ", "ผู้ทำรายการ", "ผลลัพธ์", "ประเภทการเข้าพัก", "เคส", "รายละเอียด"}}
	for _, item := range data.Audits {
		rows = append(rows, []any{
			item.ID,
			item.Timestamp.In(loc).Format(time.RFC3339),
			item.RoomID,
			item.Action,
			item.Actor,
			item.Result,
			item.StayMode,
			item.CaseCode,
			item.Detail,
		})
	}
	return writeRows(f, sheet, rows)
}

func writeRoomsSheet(f *excelize.File, data store.DailyWorkbookData, loc *time.Location) error {
	sheet := "Rooms"
	f.NewSheet(sheet)
	rows := [][]any{{"ห้อง", "ชั้น", "สถานะ", "ประเภทการเข้าพัก", "เคส", "คอนโทรลเลอร์ออนไลน์", "นาทีคงเหลือ", "ยอดชำระล่าสุด", "วิธีชำระล่าสุด", "เลขอ้างอิงล่าสุด", "อัปเดตล่าสุด"}}
	for _, item := range data.Rooms {
		rows = append(rows, []any{
			item.RoomID,
			item.Floor,
			item.Status,
			item.StayMode,
			item.CaseCode,
			item.ControllerOnline,
			item.RemainingMinutes,
			item.LastPaymentAmount,
			item.LastPaymentMethod,
			item.LastPaymentRef,
			item.UpdatedAt.In(loc).Format(time.RFC3339),
		})
	}
	return writeRows(f, sheet, rows)
}

func writeRows(f *excelize.File, sheet string, rows [][]any) error {
	cols, err := columnCount(rows)
	if err != nil {
		return err
	}

	for rowIdx, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, rowIdx+1)
		if err != nil {
			return fmt.Errorf("calculate excel cell: %w", err)
		}
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			return fmt.Errorf("write sheet %s row %d: %w", sheet, rowIdx+1, err)
		}
	}

	if len(rows) > 0 && cols > 0 {
		if err := styleSheet(f, sheet, rows, cols); err != nil {
			return err
		}
	}

	if len(rows) > 0 {
		if err := f.SetPanes(sheet, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		}); err != nil {
			return fmt.Errorf("freeze pane for %s: %w", sheet, err)
		}
	}
	if cols > 0 {
		lastCol, err := excelize.ColumnNumberToName(cols)
		if err != nil {
			return fmt.Errorf("column name for %s: %w", sheet, err)
		}
		if err := f.SetColWidth(sheet, "A", lastCol, 22); err != nil {
			return fmt.Errorf("set width for %s: %w", sheet, err)
		}
	}
	return nil
}

func styleSheet(f *excelize.File, sheet string, rows [][]any, cols int) error {
	lastCol, err := excelize.ColumnNumberToName(cols)
	if err != nil {
		return fmt.Errorf("column name for style %s: %w", sheet, err)
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 15, Color: "1F2937"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E7EEF8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	if err != nil {
		return fmt.Errorf("create title style for %s: %w", sheet, err)
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2F5D62"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return fmt.Errorf("create header style for %s: %w", sheet, err)
	}

	if len(rows[0]) == 1 && len(rows) > 1 {
		if cols > 1 {
			if err := f.MergeCell(sheet, "A1", lastCol+"1"); err != nil {
				return fmt.Errorf("merge title row for %s: %w", sheet, err)
			}
		}
		if err := f.SetCellStyle(sheet, "A1", lastCol+"1", titleStyle); err != nil {
			return fmt.Errorf("style title row for %s: %w", sheet, err)
		}
		return nil
	}

	if err := f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle); err != nil {
		return fmt.Errorf("style header row for %s: %w", sheet, err)
	}
	return nil
}

func columnCount(rows [][]any) (int, error) {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	return maxCols, nil
}

func formatTime(value *time.Time, loc *time.Location) string {
	if value == nil {
		return ""
	}
	return value.In(loc).Format(time.RFC3339)
}
