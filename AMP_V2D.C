/********************************************************************/
/*	Project 			Auto Switch Control 16 Output				*/
/*	Hardware/Version	ASCM version 1.0							*/
/*	Software/Version	ASCMH1.C    								*/
/*	CPU 				89C51RD2 + 64 Kbyte							*/
/*	Memory				62256 32	Kbyte							*/
/*	X-tal Freq			11.0592 MHz 								*/
/*	Customer Name		Amphan Technology							*/
/*	Customer Tel/Fax	02 9285342-3								*/
/*	Software Eng.		Kadit Onchum								*/
/*	Start  Date 		10-06-2000									*/
/*  Last Modify Date    26-07-2002                                  */
/*	Status				Modify										*/
/*  Remark              Linked witn LED For Nakhornsri Version      */
/*  					Address of LED Inidicator = 0xfe            */
/*  					Address of PRINTER        = 0xfd            */
/*						For Special Hotel							*/
/********************************************************************/

#include <reg52.h>						/* 8051 Register	  */
#include <stdio.h>						/* For standard I/O   */
#include <absacc.h> 					/* Absolute access	  */
#include <ctype.h>						/* Character type	  */
#include <stdlib.h> 					/* Standard library   */
#include <intrins.h>					/* Intrinstric func   */

#define LCDWRC		XBYTE[0XA000]		/* LCD write command */
#define LCDRDC		XBYTE[0XA001]		/* LCD read  command */
#define LCDWRD		XBYTE[0XA002]		/* LCD write data	  */
#define LCDRDD		XBYTE[0XA003]		/* LCD read  data	  */

#define PRNDAT		XBYTE[0X8000]		/* Printer data	  */
#define PRNSTB		XBYTE[0X8001]		/* 7= /stb 		  */
#define PRNBSY		XBYTE[0X8002]		/* 7= busy 		  */
#define CON8255 	XBYTE[0X8003]		/* Control port	  */
#define KEYCOL		XBYTE[0X8001]		/* 0-5 = key col	  */
#define KEYROW		XBYTE[0X8002]		/* 0-4 = key row	  */

#define RTCADDR 	0xd0			   	/* DS1307 addressing	 */
#define DISPAD		0xfe			   	/* Display board Address */
#define WDS 		(WDSBIT = ~WDSBIT) 	/* Watchdog trigger 	 */
#define CARD		16					/* 1 Slave Maximum		 */
#define PORT		16					/* 1 Slave 16 Port		 */	 

sfr 	AUXR	=	0x8e;			   // For 89C51RD+ access xram 

sbit	LEDRX	=	P1^1;			   // RX LED		
sbit	LEDTX	=	P1^0;			   // TX LED		
sbit	SOUNDB	=	P1^4;			   // Sound bit LED 
sbit	SCL 	=	P1^5;			   // SCL			
sbit	SDA 	=	P1^6;			   // SDA			
sbit	WDSBIT	=	P1^7;			   // Watch dog 	
sbit	CON485	=	P3^4;			   // 485 control	

#define KEYDEL1 	800
#define KEYDEL2 	400
#define UP			'U'
#define DN			'D'
#define PROG1		'A'
#define PROG2		'B'
#define IN			'I'
#define OUT 		'O'
#define CLS 		'C'
#define ENT 		'E'
#define DATEP		'+'
#define DATES		'-'
#define PRINT		'P'
#define STOP		'S'
#define TX			1
#define RX			0

#define TIM0HI		0xf1			   /* Timer-0 high byte data 0xf8cd 1 mSec */
#define TIM0LO		0x9a			   /* Timer-0 low byte data  0xf19a 2 mSec */
#define MAXREC		3800			   /* Maximum records 2000 */

#define LEDOff		0			   		// Led Off
#define LEDSlow		1			   		// Led Slow
#define LEDFast		2			   		// Led Fast
#define LEDOn		3			   		// Led On

#define STANone		0			   		// Guest Not Use
#define STAout		1			   		// Guest Out
#define STAin		2			   		// Guest In
#define STAcup		3			   		// Guest Clean
#define STAot		4			   		// Guest Ot

void SetRoom(void);
void SetSlave(void);
void SetupMachine(void);
void SetSlaveRoom(void);
void SetRTC(void);
void SetTime(void);
char SetDigit(unsigned char *digit,char pos,char max1,char max2,char dg);
void SetFunction(void);
void SetFunction1(void);
void SetPassword(void);
void SetCupTime(void);
void SetOThour(void);
int  SetOtime(void);
void ChangePerion(void);
void ClearPeriod(void);
void ListPeriod(void);
void ListRecord(void);
void SendBlink(char led,unsigned char mode);
void InitRoom(void);

void WakeupSlave(void);
void Delay(unsigned int delaytime);
bit  GetPassword(char addr);
int  GetOtime(int room);
void InitRTC(void);
void GetRTC(void);
char ScanKey(void);
void ListRoom(void);
void RoomOnoff(int room,char sw);
//void SaveStatus(int room,char mode,char date,char hour,char min);
void SaveStatus(int room,unsigned char mode,char hour,char min);
void SaveCuptime(int room);
unsigned char GetStatus(int room,char *h,char *m);
void UpdateLED_All(void);
void UpdateLED(int room,char sts);

/*** PRINTER FUNCTION *****/
void PrnRec(char Sum);
void PrnInfo(void);
void PrintDateMonth(char startDate,char endDate,char startMonth,char endMonth);
void PrintSum(char startDate,char endDate,char startMonth,char endMonth);
void PrnError(void);
void PrnChar(char c);
void PrnString(char *s);
void PrnForm1(void);
void PrnForm2(void);
//void PrnRecord(int rec,char gin,char gout,char cup);
void PrnRecord(int rec,unsigned char mode,unsigned char out);
void PrnBlank(char blank);
void PrnPeriod(void);

/*** I2C Bus Functions *****/
bit  Putreg(char addr,char reg,char dat);
bit  Getreg(char addr,char reg,char *dat);
bit  Txbyte(unsigned char com);
char Rxbyte(void);
void Start(void);
void Stop(void);
void Plscl(void);
void Setscl(void);
void Clrscl(void);
void Bitdl(void);

/*** LCD Functions ******/
char putchar(char c);		   // Put 1 character to LCD		
void LcdClear(void);		   // Clear LCD 					
void LcdWait(void); 		   // Wait LCD to ready 			
void Lcdwd(char dat);		   // Write 1 data to LCD			
void Lcdwi(char com);		   // Write 1 instruction to LCD	
void LcdPos(char pos);		   // Set LCD position				
void LcdSet(void);			   // Initialize LCD				

/*** Sound Genetaror Functions ***/
void Beep(char freq,int time); // Beep sound generator function 
void PowerBeep(void);		   // Power Beep					
void ErrBeep(bit dly);		   // Error Beep					
void OverBeep(bit dly); 	   // Over range Beep				
void UnderBeep(bit dly);	   // Under range Beep				
void SuccBeep(bit dly); 	   // Success Beep					
void NormBeep(bit dly); 	   // Normal Beep					

void ClrDisbuf(void);
void PrnFunc(void);
void ToPrint(void);
void Tx485(char ser);

void SelectRoom(char *ind,char dat);
void ClearRam(void);
void CleanUP(void);
void CheckOverTime(void);
void CheckNormTime(void);
bit  CheckPrinter(void);

int  TimetoInt(char hour,char min);
void DispStatusIN(char index);
void DispStatusOUT(char index);
void SearchSlave(char *card,char *port,int room);

void ListFunction(void);
void SaveAndPrint(int room,unsigned char mode,unsigned char out);
void SaveOtime(int room,int ot);

void Out485_16(unsigned char addr,unsigned int dat);
void SendSerial(char ser);
void SetAlarm(char rem);
void ClearRecord(void);
void ClearRecCompleted(void);
void Direct_Prn(char day);
void Set_Imm_Prn(void);

typedef struct TIME{
	unsigned char mode;
	char hour;
	char min;
	int  totalmin;
}TIME;

typedef struct CLOCK {
	char sec;
	char min;
	char hour;
	char day;
	char date;
	char month;
	char year;
}CLOCK;

typedef struct GUEST {
	int  room;
	unsigned char mode :4;
	unsigned char out :4;
	char year;
	char month;
	char date;
	char hour;
	char min;
}GUEST;   

/*** External Ram Array Used *****/		
xdata CLOCK RTC;					  		// RTC buffer
xdata GUEST RECORD[MAXREC]; 			 	// Raw data record 	
xdata TIME  TSTATUS[CARD][PORT]; 			// Slave,Port Time & Mode
xdata int 	ROOMOUT[CARD];				 	// Room Output
xdata int 	ROOM[CARD][PORT];			 	// Room Number     
xdata char 	DISPBUF[42]; 				 	// Display buffer		   
xdata char 	PASSWORD[4]; 				 	// Password Buffer 	   

/*** External Ram Used *****/
xdata int 	PrintRef;                       	// Print Referrence      
xdata int  	RecordIndex;					// Period Index		   
xdata int  	PeriodStart;					// Period Start		   
xdata int  	PeriodEnd;					 	// Period End			   
xdata char 	NUMCARD;					  					   
xdata char 	CUPTIME; 					 	// Clean up time		   
xdata char 	PowMem;						 	// Memory Status			   
xdata int 	OTHOUR;						 	// Over Time Hour		   
xdata char 	REMAIN1;
xdata char 	REMAIN2;
xdata char 	IMM_DAY;

/*** Internal Ram Global Used ****/
char LCDPOS;
int  KeyCnt;
bit  Keyflag;
bit  Msecflag;
bit  Dspflag;							// Timer flag
bit  Prnflag;							// Printer flag

/*** Start Main Function *******************/
void main(void) {
	bit dispfg,timefg;
	unsigned char dispcnt;
	char key,oldmin;
	char index,sect;

	AUXR  = 0x02;						// 89c51rd2 
	SCON  = 0x52;
	TMOD  = 0x20;
	TH1   = 0xfa;       				// 0xf4 = 4800 bps 0xfa = 9600 bps 
	TH1   = 0xfa;
	TR1   = 1;
	TMOD |= 0x01;
	TH0   = TIM0HI;
	TL0   = TIM0LO;
	EA	  = 1;
	SCL   = 1;
	SDA   = 1;
	CON8255 = 0x89;
	InitRTC();
	LcdSet();
	PowerBeep();

	if(PowMem != 0x55) { 	  			// Check first start 
		ClearRam(); 		  			// Clear ram if machine is first start up 
		InitRoom();
		PowMem = 0x55;
	}

	if(ScanKey() == PROG1) {  			// P1 Pressed    
		SetupMachine(); 	  			// Setup machine 
	}

	LcdClear(); 					  	// Clear LCD			  
	//printf("AUTO SWITCH CONTROL\n");  	// Display machine name 
    printf(" AMPHAN TECHNOLOGY\n");  	// Display machine name 
    printf(" V2.0  01/12/2013");     	// Display version      
	Delay(2000);					  	// Delay 2 sec		  

	LcdClear(); 					  	// Clear LCD				 
	printf("   Ckeck Printer\n");      	// Display message         
    Prnflag = CheckPrinter();         	// Check printer installed 
	Delay(500);
    if(Prnflag) {
        printf("  Printer Connected"); 	// Printer OK         
        SuccBeep(0);
    }
    else {
        printf("   Printer Error");    	// Printer error      
        ErrBeep(0);
    }
	Delay(1000);						// Delay 1.5 sec 	 

	CON485 = TX;					  	// Control RS-485 to Transmit 
	WakeupSlave();					  	// Wake up slave (room)		
	UpdateLED_All();				  	// Update all led display 	
	PowerBeep();					  	// Power up beep sound		
	LcdClear(); 					  	// Clear display				

	dispcnt = 1;
	index	= 0;
	sect	= 0;
	dispfg	= 0;
	timefg	= 0;
	while(1) {
		WDS;
		if (Dspflag){
			Dspflag = 0;
			if(--dispcnt <= 0) {		   			// 250 mSec
				GetRTC();
				if(timefg) {							// Show Time and Date
					dispcnt = 125;
					LcdPos(0);
					printf("  AMPHAN TECHNOLGY\n");
					printf("%02bx:%02bx:%02bx  %02bx/%02bx/20%02bx",RTC.hour,RTC.min,
					RTC.sec,RTC.date,RTC.month,RTC.year);
				}
				else {
					dispcnt = 250;
					dispfg	= ~dispfg;					// Show Room
					if(dispfg) DispStatusIN(index);	  	// Room Number Index
					else	   DispStatusOUT(index);	// Room Number Index
					LcdPos(0);
					printf(DISPBUF);
				}
			}
		}

		key = ScanKey();   		
        if(key == '#') {						// Switch Show Time and Date / Room
			timefg = ~timefg;
			SuccBeep(1);
			LcdClear();
			dispcnt = 1;
		}

		else if(key == UP) {					// Change display page up 
			if(++index > ((NUMCARD * 2) -1)) {
				OverBeep(1);
				index = 0;
			}
			ClrDisbuf();
			LcdClear();
			dispcnt = 1;
		}

		else if(key == DN) {					// Move page down			
			if(--index < 0) {
				UnderBeep(1);
				index = ((NUMCARD * 2) -1);
			}
			ClrDisbuf();
			LcdClear();
			dispcnt = 1;
		}

		else if(key == PROG1) { 				// To edit			
			SetFunction();                       			
			LcdClear();
		}

		else if(key == PROG2) { 				// Change periode 
			LcdClear();
			printf("\n   PASSWORD [xxxx]");
			if(GetPassword(33)) {
				SetFunction1();
			}
			LcdClear();
		}

		else if(key == PRINT) { 				// To print  
			LcdClear();
			printf("\n   PASSWORD [xxxx]");
			if(GetPassword(33)) {
				if(Prnflag) PrnFunc();
				else {
					LcdClear();
	                printf("    Printer Error");
					Delay(1000);
	                ErrBeep(0);
				}
				LcdClear();
			}
		}

		else if(key == '+') {
        	LcdClear();
			printf(" Immediately Print");
			if (Prnflag) Direct_Prn(IMM_DAY);
			else {
				LcdClear();
	               printf("    Printer Error");
				Delay(1000);
	               ErrBeep(0);
			}
			LcdClear();
		}

		else if(isdigit(key)) { 		// Key '0'  to '9'           
			SelectRoom(&index,key);
			timefg = 0;
			LcdClear();
		}

		if(oldmin != RTC.min) { 		// Second change			 
			LEDRX = 0; 					// On LED					 
			oldmin = RTC.min;			// Update old second		 
			Prnflag  = CheckPrinter();	// Check printer every minut 
            CleanUP();                  // Check clean up time       
			CheckOverTime();			// Check over time alarm/out 
			LEDRX  = 1; 				// Off LED					 
		}
	}
}
/*** End of main function *****************/

void SetFunction1(void) {/*"--------------------"*/
	char code *menu[] = {{" 1.Add Period"},
						 {" 2.Clear Period"},
						 {" 3.List Period"},
						 {" 4.List Record"},
						 {" "}};
	char key,ind;
	bit  dispfg;
	dispfg = 1;
	ind    = 0;
	while(1) {
		WDS;
		if(dispfg) {
			dispfg = 0;
			LcdClear();
			printf(menu[ind]);
			LcdPos(20);
			printf(menu[ind+1]);
			LcdPos(0);
			putchar('>');
		}
		key = ScanKey();
		if(key == UP) {
			dispfg = 1;
			if(++ind > 3) {
				ind = 3;
				OverBeep(1);
			}
		}
		else
		if(key == DN) {
			dispfg = 1;
			if(--ind < 0) {
				ind = 0;
				UnderBeep(1);
			}
		}
		else
		if(key == ENT) {
			dispfg = 1;
			switch(ind) {
				case 0: ChangePerion(); break;
				case 1: ClearPeriod(); break;
				case 2: ListPeriod(); break;
				case 3: ListRecord(); break;
			}
		}
		else
		if(key == STOP) {
			LcdClear();
			SuccBeep(0);
			return;
		}
	}
}

void ChangePerion(void) {  /* Change Period */
	char key;
	LcdClear();
	printf("    Add Period\n");
	printf("   PASSWORD [xxxx]");
	while(1) {
		WDS;
		key = ScanKey();
		if(key == ENT) {
			if(GetPassword(33)) {
				/*if ((PeriodStart != PeriodEnd) || (RecordIndex == PeriodEnd)){
					LcdClear();
					printf("  Not Add Period");
					ErrBeep(1);
					return;
				}*/
				//PeriodStart = PeriodEnd;
				PeriodEnd = RecordIndex;
				LcdClear();
				printf("     Add Period\n");
				printf("     Completed");
				Delay(1000);
				SuccBeep(0);
				return;
			}
		}
		else if(key == STOP) {
			ErrBeep(1);
			return;
		}
	}
}

void ClearPeriod(void) {  /* Clear Period */
	int i,j;
	char key;
	LcdClear();
	printf("    Clear Period\n");
	printf("   PASSWORD [xxxx]");
	while(1) {
		WDS;
		key = ScanKey();
		if(key == ENT) {
			if(GetPassword(33)) {
				if (PeriodStart == PeriodEnd) {
					LcdClear();
					printf("     No Period");
					ErrBeep(1);
					return;
				}
				for(i=0,j=PeriodEnd;j<RecordIndex;i++,j++) {
					RECORD[i].room  = RECORD[j].room;
	   				RECORD[i].date  = RECORD[j].date;
	   				RECORD[i].month = RECORD[j].month;
	   				RECORD[i].year  = RECORD[j].year;
	   				RECORD[i].mode  = RECORD[j].mode;
					RECORD[i].out   = RECORD[j].out;
	   				RECORD[i].hour  = RECORD[j].hour;
	   				RECORD[i].min   = RECORD[j].min;
				}
				RecordIndex = i;
				PeriodStart = 0;
				PeriodEnd = 0;
				for(;i<MAXREC;i++) {
					RECORD[i].room  = 0;
	   				RECORD[i].mode  = 0;
					RECORD[i].out   = 0;
				}
				LcdClear();
				printf("    Clear Period\n");
				printf("     Completed");
				Delay(1000);
				SuccBeep(0);
				return;
			}
		}
		else if(key == STOP) {
			ErrBeep(1);
			return;
		}
	}
}

void ListPeriod(void) {  /* List Period */
	bit dispfg;
	int ind;
	char key;
	unsigned char mode;
	LcdClear();
	printf("    List Period\n");
	Delay(500);
	while(1) {
		WDS;
		//key = ScanKey();
		//if(key == ENT) {
			if (PeriodStart == PeriodEnd) {
				LcdClear();
				printf("     No Period");
				ErrBeep(1);
				return;
			}
			ind = 0;
			dispfg = 1;
			while(1){
				WDS;
				if (dispfg){
					LcdClear();
					printf("ROOM DD/MM HH:MM STS\n");
					printf(" %03d",RECORD[ind].room);  							// print room    
					printf(" %02bx/%02bx",RECORD[ind].date,RECORD[ind].month);	// print date/month
					printf(" %02bx:%02bx",RECORD[ind].hour,RECORD[ind].min);	// print hour:min
					mode = RECORD[ind].mode;
					if (mode == STAout) printf(" OUT");							// print Out mode
					else if (mode == STAin) printf(" IN");						// print In mode
					else if (mode == STAcup) printf(" CUP");					// print Cup mode
					else if (mode == STAot) printf(" OT");						// print Ot mode
					dispfg = 0;
				}

				key = ScanKey();
				if(key == STOP) {	// Exit 
					SuccBeep(0);
					return;
				}
				else
				if(key == UP) {
					dispfg = 1;
					if(++ind >= PeriodEnd) {
						ind = PeriodEnd-1;
						OverBeep(1);
					}
				}
				else
				if(key == DN) {
					dispfg = 1;
					if(--ind < PeriodStart) {
						ind = PeriodStart;
						UnderBeep(1);
					}
				}
			}
		//}
		//else if(key == STOP) {
		//	ErrBeep(1);
		//	return;
		//}
	}
}

void ListRecord(void) {  /* List Period */
	bit dispfg;
	int ind;
	char key;
	unsigned char mode;
	LcdClear();
	printf("    List Record\n");
	Delay(500);
	while(1) {
		WDS;
		//key = ScanKey();
		//if(key == ENT) {
			if (RecordIndex == 0) {
				LcdClear();
				printf("     No Record");
				ErrBeep(1);
				return;
			}
			ind = 0;
			dispfg = 1;
			while(1){
				WDS;
				if (dispfg){
					mode = RECORD[ind].mode;
					LcdClear();
					printf("ROOM DD/MM HH:MM STS\n");
					printf(" %03d",RECORD[ind].room);  							// print room    
					printf(" %02bx/%02bx",RECORD[ind].date,RECORD[ind].month);	// print date/month
					printf(" %02bx:%02bx",RECORD[ind].hour,RECORD[ind].min);	// print hour:min
					if (mode == STAout) printf(" OUT");							// print Out mode
					else if (mode == STAin) printf(" IN");						// print In mode
					else if (mode == STAcup) printf(" CUP");					// print Cup mode
					else if (mode == STAot) printf(" OT");						// print Ot mode
					else printf(" ");
					dispfg = 0;
				}

				key = ScanKey();
				if(key == STOP) {	// Exit 
					SuccBeep(0);
					return;
				}
				else
				if(key == UP) {
					dispfg = 1;
					if(++ind >= RecordIndex) {
						ind = RecordIndex-1;
						OverBeep(1);
					}
				}
				else
				if(key == DN) {
					dispfg = 1;
					if(--ind < 0) {
						ind = 0;
						UnderBeep(1);
					}
				}
			}
		//}
		//else if(key == STOP) {
		//	ErrBeep(1);
		//	return;
		//}
	}
}

void CheckOverTime(void) {
    char i,j;
	int  room;
	int  rtime,otime,tmp;
    for(i=0; i<NUMCARD; i++) {                 								// Card   loop              
		for(j=0; j<PORT; j++) {					 							// Output loop 			 	 
			if(TSTATUS[i][j].mode == STAot) { 		 						// Over   time
                rtime = TimetoInt(RTC.hour,RTC.min);         				// Get real time in integer                 
				otime = TimetoInt(TSTATUS[i][j].hour,TSTATUS[i][j].min);
                if((rtime - otime) < 0) {  									// Real Time - Over Time 
                    rtime += (int)(1440);
                }
				tmp = TSTATUS[i][j].totalmin;
                if((rtime - otime) >= tmp) {
					room = ROOM[i][j];
					SaveAndPrint(room,STAout,STAot); 						// out
					RoomOnoff(room,STAout);					 				// Off 	 
					UpdateLED(room,LEDOff);					 				// LED off  
				}
				else if((rtime - otime) >= (tmp - REMAIN2)) {
					UpdateLED(ROOM[i][j],LEDFast);			 				// LED Fast 
				}
				else if((rtime - otime) >= (tmp - REMAIN1)) {
					UpdateLED(ROOM[i][j],LEDSlow);			 				// LED Slow 
				}
			}
		}
	}
}

void SelectRoom(char *ind,char dat) {
	char key,index;
	unsigned char mode,temp;
	char hour,min;
	char card,port;
	char roombuf[4] = {0,0,0,0};
	int  room,ot,tout;
	//int  rtime,ctime;
	index = 0;
	LcdClear();
	printf("  ROOM NUMBER [   ]\n");
	printf("STATUS [  :  ][   ]");
	LcdPos(15);
	Lcdwi(0x0e);			// Lcd bllink command 
	putchar(dat);
	roombuf[index] = dat;
	tout = 0;
	while(1) {
		WDS;
        if(++tout > 20000) {
			ErrBeep(0);
			return;
		}
		key = ScanKey();
		if(key == STOP) {	// Exit 
			ErrBeep(1);
			return;
		}
		else
		if(isdigit(key)) {
			tout = 0;
			if(++index < 2) {
				putchar(key);
				roombuf[index] = key;
			}
			else {
				putchar(key);
				roombuf[index] = key;
				Lcdwi(0x0c);
				room = atoi(roombuf);
				mode = GetStatus(room,&hour,&min);
				/*if((mode == STAin) || (mode == STAcup) || (mode == STAot)) {
					ctime = TimetoInt(hour,min);
					rtime = TimetoInt(RTC.hour,RTC.min);
					if((rtime - ctime) < 0) {
						rtime = (rtime + 1440);
					}
        			rtime = (rtime - ctime);
          			hour = (char)(rtime/60);
          			min  = (char)(rtime%60);
				}*/
					
				LcdPos(35);
				switch(mode) {
					case STANone: {			// No port define 
						LcdPos(27);
						printf("[  :  ]");
						LcdPos(34);
						printf("[ NO]");
						Delay(1000);
						ErrBeep(0);
						return;
					}
					case 0xff:
					case STAout: printf("OUT"); break;
					case STAin: printf(" IN"); break;
					case STAcup: printf("CUP"); break;
					case STAot: printf(" OT"); break;
				}
				if (mode > STAout){
					LcdPos(28);
					printf("%02bx:%02bx",hour,min);
				}
				tout = 0;
				temp = mode;
				while(1) {
					WDS;
                    if(++tout > 20000) {
						ErrBeep(0);
						return;
					}
					key = ScanKey();
					if(key) tout = 0;
					LcdPos(35);
					switch(key) {
						case STOP: ErrBeep(1); LcdClear(); return;
						case  OUT: temp = STAout; printf("OUT");  break;
						case   IN: temp = STAin; printf(" IN");  break;
						case  CLS: temp = STAcup; printf("CUP");  break;
						case  '+': temp = STAot; printf(" OT");  break;
						case  '-': {             		// ot time expan            
							if(mode == STAot) {
								ot = SetOtime(); 		// Setting and return value 
								if(ot > 0) {
									printf("\nADD OT %d MIN\n",ot);
									Delay(1000);
									SuccBeep(0);
									ot += GetOtime(room);
									SaveOtime(room,ot);
									UpdateLED(room,LEDOn);	// LED on
									SearchSlave(&card,&port,room);
									*ind = card * 2;
									if(port >= 8) *ind = (*ind + 1);
									return;
								}
								else {
									ErrBeep(0);
									return;
								}
							}
							else {
								ErrBeep(1);
								return;
							}
						}
						case  ENT: {
							if(mode != temp) {
								if (((mode == STAin) || (mode == STAot) || (mode == STAcup)) &&
									((temp == STAin) || (temp == STAot) || (temp == STAcup))){
									ErrBeep(1);
									return;	
								}
								switch (temp){
									case STAout:	SaveAndPrint(room,temp,mode);
													UpdateLED(room,LEDOff);  // LED off 
													break;
									case STAot:		SaveOtime(room,OTHOUR);
									case STAin:
									case STAcup:	SaveAndPrint(room,temp,0);
													UpdateLED(room,LEDOn);  // LED on	
													break;
								}
								RoomOnoff(room,temp);
								SearchSlave(&card,&port,room);
								*ind = card * 2;
								if(port >= 8) *ind = (*ind + 1);
								SuccBeep(1);
								return;
							}
							else {						   			// Status not change 
								ErrBeep(1);
								return;
							}
						}
					}
				}
			}
		}
	}
}

bit CheckPrinter(void) {
	char  p;
	int   i;
	PRNSTB = p | 0x80;
	for(i=0;i<100;i++);

	for(i=0;i<100;i++) {			 				// Check Printer 100 ms 
		if(!(PRNBSY & 0x80)) return(1);				// Wait for Busy = 0
		Delay(2);									// Delay 2 ms 
	}
	return(0);
}

void SetupMachine(void) {
char code *Set_menu[] = {{" Set Password"},        	// 0 
						 {" Set Max Slave"},       	// 1 
						 {" Set clean up time"}, 	// 2 
						 {" Set OT Hour"},         	// 3 
						 {" Set Alarm-1"},			// 4
						 {" Set Alarm-2"},          // 5
						 {" Set Immediately prn"},  // 6
						 {" Clear All Record"},     // 7 
						 {" Clear Ram"},           	// 8 
						 {" "}};//Initialize Room"}};

	char key,ind;
	bit dispfg;
	dispfg = 1;
	ind = 0;
	while(1) {
		WDS;
		if(dispfg) {
			dispfg = 0;
			LcdClear();
			printf("%s",Set_menu[ind]);
			LcdPos(20);
			printf("%s",Set_menu[ind+1]);
			LcdPos(0);
			putchar('>');
		}

		key = ScanKey();
		if(key == UP) {
			dispfg = 1;
			if(++ind > 8) {
				ind = 8;
				OverBeep(1);
			}
		}
		else
		if(key == DN) {
			dispfg = 1;
			if(--ind < 0) {
				ind = 0;
				UnderBeep(1);
			}
		}
		else
		if(key == ENT) {
			dispfg = 1;
			switch(ind) {
				case 0: SetPassword(); break;
				case 1: SetSlave();    break;
				case 2: SetCupTime();  break;
				case 3: SetOThour();   break;	  // Over Time Hour   
				case 4: SetAlarm(0);   break;
				case 5: SetAlarm(1);   break; 
				case 6: Set_Imm_Prn(); break;
				case 7: ClearRecord();
						ClearRecCompleted(); break;
				case 8: ClearRam();
						InitRoom();    break;	  // Clear  Ram 
			}
		}
		else if(key == STOP) {
			ErrBeep(1);
			return;
		}

        else if(key == '#') {
			InitRoom();
			dispfg = 1;
		}
	}
}

void InitRoom(void) {
	char i,j;
	int room;
	LcdClear();
	printf("  Initialize Room");
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			room = ((i+1) * 100) + (j+1);
			ROOM[i][j] = room;
		}
	}
	Delay(2000);
	SuccBeep(1);
}

void DispStatusIN(char index) {
	char mode,addr,slave;
	char i,j,start,end;
	int  room;
	slave = (index/2);
	if(index % 2) {			// OP 9-16	
		start = (PORT/2);
		end = PORT;
	}
	else {				// OP 1-8 
		start = 0;
		end = (PORT/2);
	}
	for(j=0,i=start; i<end; i++) {
		room = ROOM[slave][i];
		mode = TSTATUS[slave][i].mode;
		if(room != 0) {
			addr = (j * 5);
			j++;
			sprintf(&DISPBUF[addr],"%03d  ",room);
			if(mode == STAin) {							// In	
				sprintf(&DISPBUF[addr]," IN  ");
			}
			else if(mode == STAot) { 					// OT
				sprintf(&DISPBUF[addr]," OT  ");
			}
			else if(mode == STAcup) {					// Clean Up 
				sprintf(&DISPBUF[addr],"CUP  ");
			}
		}
	}
	DISPBUF[39] = '\0';
}

void DispStatusOUT(char index) {
	char i,j,start,end,slave;
	int  room;

	slave = (index/2);
	if(index % 2)  { 		// OP 9-16 
		start = PORT/2;
		end = PORT;
	}
	else {				// OP 1-8 
		start = 0;
		end = PORT/2;
	}
	for(i=0; i<40; i++) DISPBUF[i] = ' ';  // Clear display buffer 
	for(j=0,i=start; i<end; i++) {
		room = ROOM[slave][i];
		if(room != 0) {
			sprintf(&DISPBUF[j*5],"%03d  ",room);
			j++;
		}
	}
	DISPBUF[39] = '\0';
}

void CleanUP(void) {
	char  i,j;
	int room;
	int   rtctime,cleantime;
	rtctime = TimetoInt(RTC.hour,RTC.min);
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(TSTATUS[i][j].mode == STAcup){
				room = ROOM[i][j];
				cleantime = TimetoInt(TSTATUS[i][j].hour,TSTATUS[i][j].min);
				if(rtctime - cleantime < 0) {
					rtctime = (rtctime + 1440);
				}

				if((rtctime - cleantime) >= CUPTIME) {
					SaveAndPrint(room,STAout,STAcup);
					RoomOnoff(room,STAout);
					UpdateLED(room,LEDOff);
				}
				else if((rtctime - cleantime) >= (CUPTIME - 1)) {
					UpdateLED(room,LEDFast);
				}
			}
		}
	}
}

/**** Convert time to integer ****/
int TimetoInt(char hour,char min) {
	char ascbuf[4];
	int  intg;
	sprintf(ascbuf,"%02bx",hour);
	intg   = atoi(ascbuf);
	intg  *= 60;
	sprintf(ascbuf,"%02bx",min);
	intg  += atoi(ascbuf);
	return(intg);
}

void SetFunction(void) {/*"--------------------"*/
	char code *menu[] = {{" 1.Set date/time"},
						 {" 2.Set slave/room"},
						 {" 3.List slave/room"},
						 {" "}};
	char key,ind;
	bit  dispfg;
	dispfg = 1;
	ind    = 0;
	while(1) {
		WDS;
		if(dispfg) {
			dispfg = 0;
			LcdClear();
			printf(menu[ind]);
			LcdPos(20);
			printf(menu[ind+1]);
			LcdPos(0);
			putchar('>');
		}
		key = ScanKey();
		if(key == UP) {
			dispfg = 1;
			if(++ind > 2) {
				ind = 2;
				OverBeep(1);
			}
		}
		else
		if(key == DN) {
			dispfg = 1;
			if(--ind < 0) {
				ind = 0;
				UnderBeep(1);
			}
		}
		else
		if(key == ENT) {
			dispfg = 1;
			switch(ind) {
				case 0: SetTime();	  break;
				case 1: SetRoom();	  break;
				case 2: ListRoom();   break;
			}
		}
		else
		if(key == STOP) {
			LcdClear();
			ErrBeep(1);
			return;
		}
	}
}

void SetCupTime(void) {
	unsigned char *ptr;
	char  buf[3];
	char  key,cup,max1,max2,dg,pos;
	LcdClear();
	printf("   SET CLEAN UP\n");
	printf("   PASSWORD [xxxx]");
	while(1) {
		WDS;
		key = ScanKey();
		if(key == STOP) {
			return;
		}
		else
		if(key == ENT) {
			if(GetPassword(33)) {
				cup = CUPTIME;
				LcdClear();
				printf("CLEAN UP [%02bd] MINUTE\n",cup);
				pos = 0;
				LcdPos(pos+10);
				Lcdwi(0x0d);
				while(1) {
					WDS;
					switch(pos) {
						case 0:  ptr = &cup;  max1='5';max2='0';dg=0; break;
						case 1:  ptr = &cup;  max1='0';max2='9';dg=1; break;
					}
					switch(SetDigit(ptr,pos+10,max1,max2,dg)) {
						case 0: Lcdwi(0x0c);return;
						case 1: {
							if(pos < 1) pos++;
							else OverBeep(1);
							break;
						}
						case 2: {
							if(pos > 0) pos--;
							else UnderBeep(1);
							break;
						}
						case 3: {
							Lcdwi(0x0c);
							sprintf(buf,"%02bx",cup);
							CUPTIME = (char)atoi(buf);
							return;
						}
					}
				}
			}
		}
	}
}

void SetPassword(void) {
	char  key,index;
	char  password[4];
	LcdClear();
	printf("ENTER NEW PASSWORD\n");
	printf("            [xxxx]");
	LcdPos(33);
	Lcdwi(0x0d);
	index = 0;
	while(1) {
		WDS;
		key = ScanKey();
		if(isdigit(key)) {
			putchar(key);
			LcdPos(33+index+1);
			password[index] = key;
			if(++index > 3) {
				Lcdwi(0x0c);
				LcdClear();
				printf("  ENTER TO CONFIRM");
				while(1) {
					WDS;
					switch(ScanKey()) {
						case STOP: return;
						case  ENT: {
							for(index=0;index<4;index++) {
								PASSWORD[index] = password[index];
							}
							SuccBeep(1);
							return;
						}
					}
				}
			}
		}
	}
}

void WakeupSlave(void) {
	char i,j;
	int  tout;
	unsigned int mask; //,dat;
	LcdClear();
	printf("ENTER to update room\n");
	printf("STOP  to cancel");
	tout = 0;
	while(1) {
		WDS;
		i = ScanKey();
		if(i == ENT) break;
		else
		if(i == STOP) return;
		if(++tout > 10000) break;
	}

	SuccBeep(0);
	SuccBeep(0);

	UpdateLED_All();
	LcdClear();
	printf(" UPDATE ROOM STATUS\n");
    printf("     [--] [--]");
	Delay(500);
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			WDS;
			if(TSTATUS[i][j].mode > STAout) {
				mask = 0xfffe;
				mask <<= j;
				mask = ~mask;
				LcdPos(26);
                printf("%02bd",i+1);
				LcdPos(31);
                printf("%02bd",j+1);
				Out485_16(i,mask);
				Delay(1000);
			}
		}
	}
}

void UpdateLED_All(void) {
	char  i,j,addr,out;
	unsigned int mask,dat;
	for(i=0; i<NUMCARD; i++) {           
		WDS;
		for(j=0; j<PORT; j++) {
			mask   = 0x0001;
			mask <<= j;
			dat  = (mask & ROOMOUT[i+1]);
			if(dat) out = LEDOn;
			else	out = LEDOff;
			addr = ((i * PORT) + j);
			SendBlink(addr,out);
		}
	}
}

void PrnFunc(void) {
	LcdClear();
	printf("1.PERIOD  2.RECORD\n");
	printf("3.SUMMARY 4.INFO");
	while(1) {
		WDS;
		switch(ScanKey()) {
			case UP:   PrnChar('\n'); break;           /* Line feed         */
			case STOP: LcdClear(); return;
			case '1': {
				PrnChar('\n');
				PrnPeriod();
				return;
			}
			case '2': PrnRec(0); return;   // Print record      
			case '3': PrnRec(1); return;   // Print Summary     
			case '4': PrnInfo(); return;   // Print Information 
		}
	}
}

void PrnPeriod(void) {
	char dspbuf[20];
	unsigned char mode;
	int room,i,j;
	bit prnfg;

	if(PeriodEnd == PeriodStart) {
		LcdClear();
		printf("    NO RECORD");
		ErrBeep(1);
		Delay(1000);
		return;
	}
	PrnChar(0x02);
	sprintf(dspbuf,"\nNo. REF: %03d\n",PrintRef);
	PrnString(dspbuf);
	sprintf(dspbuf,"Time   : %02bx:%02bx %02bx/%02bx/20%02bx\n",RTC.hour,RTC.min,RTC.date,RTC.month,RTC.year);
	PrnString(dspbuf);
	PrnForm2();                           							// print head form 
	for(room=100; room<=999; room++) {
		for(i = PeriodStart; i < PeriodEnd; i++) {
			WDS;
			if(room == RECORD[i].room) {
				mode = RECORD[i].mode;
				if((mode == STAin) || (mode == STAot) || (mode == STAcup)) { // Guest in + ot + cup
					sprintf(dspbuf,"%03d",RECORD[i].room);  	// Print room    
					PrnString(dspbuf);
					PrnBlank(3);
					
					sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[i].date,RECORD[i].month,RECORD[i].year); // print date/month/year
					PrnString(dspbuf);          					
					PrnBlank(2);
					
					sprintf(dspbuf,"%02bx:%02bx",RECORD[i].hour,RECORD[i].min);
					PrnString(dspbuf);
					PrnBlank(2);
					
					prnfg = 0;
					for(j=i; j < PeriodEnd; j++) {
						WDS;
						if(room == RECORD[j].room) {
							if(RECORD[j].mode == STAout) {
								sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[j].date,RECORD[j].month,RECORD[j].year); // print date/month/year
								PrnString(dspbuf);          					
								PrnBlank(2);

								sprintf(dspbuf,"%02bx:%02bx",RECORD[j].hour,RECORD[j].min);
	                            PrnString(dspbuf);   								// print guest out time 
	                            PrnBlank(3);
								
								if (mode == STAin){
									PrnString("In\n");
								}
								else if (mode == STAot){
									PrnString("Ot\n");
								}
								else if (mode == STAcup){
									PrnString("Cup\n");
								}
								else PrnString("\n");
								prnfg = 1;
								break;
							}
						}
					}
					if(!prnfg) {
						PrnString("     -    ");
						PrnBlank(2);
						PrnString("  -  ");
						PrnBlank(3);
						if (mode == STAin){
							PrnString("In\n");
						}
						else if (mode == STAot){
							PrnString("Ot\n");
						}
						else if (mode == STAcup){
							PrnString("Cup\n");
						}
						else PrnString("\n");
					}
				}
				else if(mode == STAout) {
					prnfg = 0;
					for(j = i; j >= PeriodStart; j--) {
						if(room == RECORD[j].room) {
							mode = RECORD[j].mode;
							if((mode == STAin) || (mode == STAot) || (mode == STAcup)) {
								prnfg = 1;
								break;
							}
						}
					}
					if(!prnfg) {
						sprintf(dspbuf,"%03d",RECORD[i].room);  // Print room    
						PrnString(dspbuf);
						PrnBlank(3);
						PrnString("     -    ");
						PrnBlank(2);
						PrnString("  -  ");
						PrnBlank(2);
						
						sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[i].date,RECORD[i].month,RECORD[i].year); // print date/month/year
						PrnString(dspbuf);
						PrnBlank(2);

						sprintf(dspbuf,"%02bx:%02bx",RECORD[i].hour,RECORD[i].min);
	                    PrnString(dspbuf);   								// print guest out time 
	                    PrnBlank(3);
						
						mode = RECORD[i].out;
						if(mode == STAin) {
							PrnString("In\n");
						}
						else if(mode == STAot) {
							PrnString("Ot\n");
						}
						else if(mode == STAcup) {
							PrnString("Cup\n");
						}
						else PrnString("\n");
					} 
				}
			}
		}
	}
	PrnString("\n");
	PrnChar(0x03);
	if(++PrintRef > 999) PrintRef = 0;
}

void PrnRec(char Sum) {
	unsigned char *ptr;
	char pos,max1,max2,dg;
	char startDate,endDate;
	char startMonth,endMonth;
	
	if(RecordIndex == 0) {
		LcdClear();
		printf("    NO RECORD");
		ErrBeep(1);
		Delay(1000);
		return;
	}

	startDate  = endDate  = RTC.date;
	startMonth = endMonth = RTC.month;
	LcdClear();
	printf("START END\n");
	printf("%02bx/%02bx %02bx/%02bx",startDate,startMonth,endDate,endMonth);
	pos = 0;
	LcdPos(pos+20);
	Lcdwi(0x0d);		 // Cursor blink 
	while(1) {
		WDS;
		switch(pos) {
			case 0:  ptr = &startDate;   max1='3';max2='0';dg=0; break;
			case 1:  ptr = &startDate;   max1='3';max2='1';dg=1; break;
			case 3:  ptr = &startMonth;  max1='1';max2='0';dg=0; break;
			case 4:  ptr = &startMonth;  max1='1';max2='2';dg=1; break;
			case 6:  ptr = &endDate;  max1='3';max2='0';dg=0; break;
			case 7:  ptr = &endDate;  max1='3';max2='1';dg=1; break;
			case 9:  ptr = &endMonth; max1='1';max2='0';dg=0; break;
			case 10: ptr = &endMonth; max1='1';max2='2';dg=1; break;
		}
		switch(SetDigit(ptr,pos+20,max1,max2,dg)) {
			case 0: Lcdwi(0x0c); return;
			case 1: {
				if(pos < 10) pos++;
                else OverBeep(1);
				if((pos==2)||(pos==5)||(pos==8)) pos++;
				break;
			}
			case 2: {
				if(pos > 0) pos--;
                else UnderBeep(1);
				if((pos==2)||(pos==5)||(pos==8)) pos--;
				break;
			}
			case 3: {
				Lcdwi(0x0c);
				LcdClear();
				printf("PRINT START  %02bx/%02bx\n",startDate,startMonth);
				printf("        END  %02bx/%02bx",endDate,endMonth);
				Delay(1000);
				if(Sum) PrintSum(startDate,endDate,startMonth,endMonth);
				else PrintDateMonth(startDate,endDate,startMonth,endMonth);\
				return;
			}
		}
	}
}

// Print Summary 
void PrintSum(char startDate,char endDate,char startMonth,char endMonth) {
	bit prnfg;
    char dspbuf[12];
	int i,j,start,end,monthdate;
	//int timein,timeout;
	unsigned char mode;

	start = (int)((startMonth*100) + startDate);
	end   = (int)((endMonth*100) + endDate);

	PrnChar(0x02);
	PrnString("\nTABLE2 PRINT SUMMARY\n");  						// print head table 
    PrnString("DATE: ");
    sprintf(dspbuf,"%02bx/%02bx",startDate,startMonth);
    PrnString(dspbuf);

    PrnString(" - ");
    sprintf(dspbuf,"%02bx/%02bx\n",endDate,endMonth);
    PrnString(dspbuf);
    PrnForm2();                           							// print head form  

    for(i=0; i<(int)MAXREC; i++) {   								// loop in record   
		WDS;
		LcdClear();
		monthdate = (int)((RECORD[i].month * 100) + RECORD[i].date);
		if((monthdate >= start) && (monthdate <= end)) {			// window detect
			mode = RECORD[i].mode;
			if((mode == STAin) || (mode == STAot) || (mode == STAcup)) {
               
				sprintf(dspbuf,"%03d",RECORD[i].room);
                PrnString(dspbuf);          					// print room no.      
				PrnBlank(3);
				
				sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[i].date,RECORD[i].month,RECORD[i].year); // print date/month/year
                PrnString(dspbuf);          					
				PrnBlank(2);

				sprintf(dspbuf,"%02bx:%02bx",RECORD[i].hour,RECORD[i].min);
                PrnString(dspbuf);          					// print guest in time 
				PrnBlank(2);

				prnfg = 0;
				for(j=i; j<(int)MAXREC; j++) {
					WDS;
					monthdate = (int)((RECORD[j].month * 100) + RECORD[j].date);
					if((monthdate >= start) && (monthdate <= end)) {			// window detect
	                	if(RECORD[i].room == RECORD[j].room) {
	                        if(RECORD[j].mode == STAout) {         					// guest out 
								
								sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[j].date,RECORD[j].month,RECORD[j].year); // print date/month/year
								PrnString(dspbuf);          					
								PrnBlank(2);

								sprintf(dspbuf,"%02bx:%02bx",RECORD[j].hour,RECORD[j].min);
	                            PrnString(dspbuf);   								// print guest out time 
	                            PrnBlank(3);
									
								/*timein = TimetoInt(RECORD[i].hour,RECORD[i].min);
								timeout = TimetoInt(RECORD[j].hour,RECORD[j].min);
	                			if((timeout - timein) < 0) {
	                    			timein = (1440 - timein);
	                    			timeout = (timeout + timein);
	                			}
	                			else {
	                    			timeout = (timeout - timein);
	                			}
	                			sprintf(dspbuf,"%02d:%02d",timeout/60,timeout%60);	 	//Used Print
	                			PrnString(dspbuf);
	                			PrnBlank(6);
								*/
								
								if (mode == STAin){
									PrnString("In\n");
								}
								else if (mode == STAot){
									PrnString("Ot\n");
								}
								else if (mode == STAcup){
									PrnString("Cup\n");
								}
								else PrnString("\n"); 
								prnfg = 1;
	                            break;	// Out from for loop
	                        }
	                    }
					}
				}
				if(!prnfg) {
					PrnString("     -    ");
					PrnBlank(2);
					PrnString("  -  ");
					PrnBlank(3);
					mode = RECORD[i].mode;
					if (mode == STAin){
						PrnString("In\n");
					}
					else if (mode == STAot){
						PrnString("Ot\n");
					}
					else if (mode == STAcup){
						PrnString("Cup\n");
					}
					else PrnString("\n");
				}
			}
			else if(mode == STAout) {
				prnfg = 0;
				for(j = i; j >= 0; j--) {
					if(RECORD[j].room == RECORD[i].room) {
						mode = RECORD[j].mode;
						if((mode == STAin) || (mode == STAot) || (mode == STAcup)) {
							prnfg = 1;
							break;
						}
					}
				}
				if(!prnfg) {
					sprintf(dspbuf,"%03d",RECORD[i].room);  // Print room    
					PrnString(dspbuf);
					PrnBlank(3);
					PrnString("     -    ");
					PrnBlank(2);
					PrnString("  -  ");
					PrnBlank(2);
						
					sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[i].date,RECORD[i].month,RECORD[i].year); // print date/month/year
					PrnString(dspbuf);
					PrnBlank(2);

					sprintf(dspbuf,"%02bx:%02bx",RECORD[i].hour,RECORD[i].min);
	                PrnString(dspbuf);   								// print guest out time 
	                PrnBlank(3);
						
					mode = RECORD[i].out;
					if(mode == STAin) {
						PrnString("In\n");
					}
					else if(mode == STAot) {
						PrnString("Ot\n");
					}
					else if(mode == STAcup) {
						PrnString("Cup\n");
					}
					else PrnString("\n");
				} 
			}
		}
	}
	PrnString("\n");
	PrnChar(0x03);
}

void PrintDateMonth(char startDate,char endDate,char startMonth,char endMonth) {
	int i,start,end,monthdate;

	start = (int)(startMonth*100) + startDate;
	end   = (int)(endMonth*100) + endDate;

	PrnChar(0x02);
	PrnString("\nTABLE1 PRINT RECORD\n");
    PrnForm1();
	for(i=0; i<MAXREC; i++) {
		WDS;
		monthdate = (int)((RECORD[i].month*100) + RECORD[i].date);
		if((monthdate >= start) && (monthdate <= end)) {
			PrnRecord(i,RECORD[i].mode,RECORD[i].out);
		}
		if(ScanKey() == STOP) {
			break;
		}
	}
	PrnString("\n");
	PrnChar(0x03);
}

void PrnInfo(void) {
	char  prnbuf[10];
	PrnChar(0x02);
	PrnString("\nAuto Switch Control System V2.0\n");
	PrnString("Amphan Technology Ltd.,Part.\n");
	PrnString("18/113 MOO 5 SONGPRAPA RD., SEEKAN ");
	PrnString("BANGKOK 10210\n");
  //PrnString("TEL.92853342-3  FAX. 9285343\n\n");
	PrnString("Tel.9617516-9  Fax. 9617520\n\n");
	PrnString("Max Slave     = ");
	sprintf(prnbuf,"%bd\n",NUMCARD);
	PrnString(prnbuf);
	PrnString("Max Record    = ");
	sprintf(prnbuf,"%d\n",MAXREC);
	PrnString(prnbuf);
	PrnString("Num Record    = ");
	sprintf(prnbuf,"%d\n",RecordIndex);
	PrnString(prnbuf);
	PrnString("Clean Up Time = ");
	sprintf(prnbuf,"%bd ",CUPTIME);
	PrnString(prnbuf);
	PrnString("Min\n");

	PrnString("OTime Hours   = ");
	sprintf(prnbuf,"%d ",OTHOUR);
	PrnString(prnbuf);
	PrnString("Min");
	PrnString("\n");
	PrnChar(0x03);
}

void SaveAndPrint(int room,unsigned char mode,unsigned char out) {

	if (RecordIndex < MAXREC){
	   	RECORD[RecordIndex].room  = room;
	   	RECORD[RecordIndex].date  = RTC.date;
	   	RECORD[RecordIndex].month = RTC.month;
	   	RECORD[RecordIndex].year  = RTC.year;
	   	RECORD[RecordIndex].mode  = mode;
		RECORD[RecordIndex].out   = out;
	   	RECORD[RecordIndex].hour  = RTC.hour;
	   	RECORD[RecordIndex].min   = RTC.min;
		
		SaveStatus(room,mode,RTC.hour,RTC.min);
		if(Prnflag) {
			PrnChar(0x02);
			PrnRecord(RecordIndex,mode,out);
			PrnChar(0x03);
		}
		RecordIndex++;
	}
	else {
		LcdClear();
		printf("    Memory Full\n");
		printf("    Clear Period");
		ErrBeep(1);
		Delay(1000); 	
   	}
}

int SetOtime(void) {
	char key,h;
	int  dat;
	h = 1;
	LcdClear();
	printf("SET OT [1.0] HOUR");
	LcdPos(8);
	while(1) {
		WDS;
		key = ScanKey();
		if(key == '1') {
			LcdPos(8);
			printf("1.0");
			h = 1;
		}
		else
		if(key == '2') {
			LcdPos(8);
			printf("2.0");
			h = 2;
		}
		else
		if(key == '3') {
			LcdPos(8);
			printf("0.5");
			h = 3;
		}
		else
		if(key == ENT) {
			Lcdwi(0x0c);
            SuccBeep(1);
			//LcdClear();
			if(h == 3) dat = 30;
            else {
                dat  = (int)(h);
                dat  = (dat * 60);
            }
			return(dat);
		}
		else
		if(key == STOP) {
            ErrBeep(1);
			return(0);
		}
	}
}

void SaveStatus(int room,unsigned char mode,char hour,char min) {
	char i,j;
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(room == ROOM[i][j]) {
			   	TSTATUS[i][j].mode = mode;
			   	TSTATUS[i][j].hour = hour;
			   	TSTATUS[i][j].min  = min;
			   	return;
			}
		}
	}
}

unsigned char GetStatus(int room,char *h,char *m) {
	char i,j;
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(room == ROOM[i][j]) {
				*h = TSTATUS[i][j].hour;
				*m = TSTATUS[i][j].min;
				return (TSTATUS[i][j].mode);
			}
		}
	}
	return 0;
}

/*void SaveCuptime(int room) {
	char i,j;
	for(i=0;i<NUMCARD;i++) {
		for(j=0;j<PORT;j++) {
			if(room == ROOM[i][j]) {
				CUPTIM[i][j].hour = RTC.hour;
				CUPTIM[i][j].min  = RTC.min;
				return;
			}
		}
	}
}*/

void SaveOtime(int room,int ot) {
	char i,j;
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(room == ROOM[i][j]) {
				TSTATUS[i][j].totalmin = ot;
				return;
			}
		}
	}
}

int GetOtime(int room) {
	char i,j;
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(room == ROOM[i][j]) {
  				return TSTATUS[i][j].totalmin;
			}
		}
	}
	return 0;
}

void SearchSlave(char *card,char *port,int room) {
	char i,j;
	for(i=0; i<NUMCARD; i++) {
		for(j=0; j<PORT; j++) {
			if(room == ROOM[i][j]) {
				*card = i;
				*port = j;
				return;
			}
		}
	}
}

void RoomOnoff(int room,char sts) {
	char slave,port;
	unsigned int out;
	for(slave=0; slave < NUMCARD; slave++) {
		for(port=0; port < PORT; port++) {
			if(room == ROOM[slave][port]) {
				out = 0x0001;
				out <<= port;
				if((sts >= STAin) && (sts <= STAot)) {
					ROOMOUT[slave] |= out;
				}
				else
				if(sts == STAout) {
					out = ~out;
					ROOMOUT[slave] &= out;
				}
				//Out485_16(slave,(ROOMOUT[slave]));
				Out485_16(slave,0xffff);
				return;
			}
		}
	}
}

void UpdateLED(int room,char sts) {
	char slave,port,out;
	for(slave=0; slave < NUMCARD; slave++) {
		for(port=0; port<PORT; port++) {
			if(room == ROOM[slave][port]) {
				out = slave;
				out = (out * PORT) + port;
				SendBlink(out,sts);
				return;
			}
		}
	}
}

void SetRoom(void) {
	char key;
	LcdClear();
	printf("   SET SLAVE/ROOM\n");
	printf("   PASSWORD [xxxx]");
	while(1) {
		WDS;
		key = ScanKey();
		if(key == STOP) {
			return;
		}
		else
		if(key == ENT) {
			if(GetPassword(33)) {
				SetSlaveRoom();
				LcdClear();
				while(1) {
					WDS;
					LcdPos(0);
					printf("ENTER  to Continue\n");
					printf("STOP   to Exit");
					key = ScanKey();
					if(key == STOP) {
						return;
					}
					else
					if(key == ENT) {
						SetSlaveRoom();
						LcdClear();
					}
				}
			}
		}
	}
}

void SetSlaveRoom(void) {
	char *ptr;
	char  asciibuf[5];
	int   room;
	char  slave,out,max1,max2,dg;
	char  room100,room1,pos;
	LcdClear();
	printf(" SLAVE PORT   ROOM\n");
	printf("  [  ].[  ] - [   ]");
	pos = 0;
	LcdPos(pos+23);
	Lcdwi(0x0d);										// Lcd blink command 
	while(1) {
		WDS;
		switch(pos) {
			case 0:  ptr = &slave;	 max1='2';max2='0';dg=0; break;
			case 1:  ptr = &slave;	 max1='2';max2='0';dg=1; break;
			case 5:  ptr = &out;	 max1='1';max2='0';dg=0; break;
			case 6:  ptr = &out;	 max1='1';max2='6';dg=1; break;
			case 12: ptr = &room100; max1='9';max2='9';dg=1; break;
			case 13: ptr = &room1;	 max1='9';max2='9';dg=0; break;
			case 14: ptr = &room1;	 max1='9';max2='9';dg=1; break;
		}
		switch(SetDigit(ptr,pos+23,max1,max2,dg)) {
            case 0: Lcdwi(0x0c); ErrBeep(1); return;  	// Exit     
			case 1: {									// Forward	
				if(pos < 14 ) pos++;
				else OverBeep(1);
				switch(pos) {
					case 2: pos = pos+3; break;
					case 3: pos = pos+2; break;
					case 4: pos = pos+1; break;

					case 7: pos = pos+5; break;
					case 8: pos = pos+4; break;
					case 9: pos = pos+3; break;
					case 10:pos = pos+2; break;
					case 11:pos = pos+1; break;
				}
				break;
			}
			case 2: {
				if(pos > 0) pos--;						// Backward 
				else UnderBeep(1);
				switch(pos) {
					case 2: pos = pos-1; break;
					case 3: pos = pos-2; break;
					case 4: pos = pos-3; break;

					case 7: pos = pos+2; break;
					case 8: pos = pos+3; break;

					case 9: pos = pos-3; break;
					case 10:pos = pos-4; break;
					case 11:pos = pos-5; break;
				}
				break;
			}
			case 3: {
				Lcdwi(0x0c);
				room100 &= 0x0f;
				sprintf(asciibuf,"%bx",room100);
				room100 = (char)atoi(asciibuf);

				sprintf(asciibuf,"%bx",room1);
				room1 = (char)atoi(asciibuf);
				room  = (int)(room100) * 100;
				room += room1;

				sprintf(asciibuf,"%bx",slave);
				slave = (char)atoi(asciibuf);

				sprintf(asciibuf,"%bx",out);
				out = (char)atoi(asciibuf);

				if((slave==0) || (out ==0)) {
					LcdClear();
					printf("SLAVE or PORT ERROR\n");
					printf("SET SLAVE = PORT = 1");
					slave = 1;
					out   = 1;
					Delay(2000);
				}
				for(pos=0;pos<PORT;pos++) {
					if(ROOM[slave][pos] == room) {
						ROOM[slave][pos] = 0;
					}
				}
				ROOM[slave][out]  = room;
				return;
			}
		}
	}
}

void SetSlave(void) {
	unsigned char *ptr;
	char  buf[3];
	char  slave,max1,max2,dg,pos=0;
	slave = NUMCARD;
	LcdClear();
	printf("MAX SLAVE [%02bd]\n",slave);
	pos = 0;
	LcdPos(pos+11);
	Lcdwi(0x0d);
	while(1) {
		WDS;
		switch(pos) {
			case 0:  ptr = &slave;	max1='1';max2='0';dg=0; break;
			case 1:  ptr = &slave;	max1='0';max2='9';dg=1; break;
		}
		switch(SetDigit(ptr,pos+11,max1,max2,dg)) {
			case 0: Lcdwi(0x0c);return;
			case 1: {
				if(pos < 1) pos++;
				else OverBeep(1);
				break;
			}
			case 2: {
				if(pos > 0) pos--;
				else UnderBeep(1);
				break;
			}
			case 3: {
				Lcdwi(0x0c);
				if (slave > 0x16) slave = 0x16;
				sprintf(buf,"%02bx",slave);
				NUMCARD = (char)atoi(buf);
				SuccBeep(1);
				return;
			}
		}
	}
}

bit GetPassword(char addr) {
	char  key,index;
	char  password[4];
	bit ok = 1;
	LcdPos(addr);
	Lcdwi(0x0d);
	index = 0;
	while(1) {
		WDS;
		key = ScanKey();
		if(isdigit(key)) {
			LcdPos(addr+index+1);
			password[index] = key;
			if(++index > 3) {
				Lcdwi(0x0c);
				for(index=0;index<4;index++) {
					if(PASSWORD[index] != password[index]) ok = 0;
				}
				if(ok) {
					LcdPos(addr);
					printf("PASS");
					SuccBeep(1);
					Delay(500);
					return(1);
				}
				else {
					ErrBeep(1);
					return(0);
				}
			}
		}
	}
}

void PrnForm1(void) {
	PrnString("==========  ====  =======  =======  ========  ========\n");
	PrnString("DATE        ROOM    IN       OUT    STATUS    REMARK  \n");
	PrnString("==========  ====  =======  =======  ========  ========\n");
}

void PrnForm2(void) {
	PrnString("====  =================  =================  ========  ========\n");
	PrnString("ROOM       GUEST IN          GUEST OUT      STATUS    REMARK  \n");
	PrnString("====  =================  =================  ========  ========\n");
}

void PrnRecord(int rec,unsigned char mode,unsigned char out) {
	char dspbuf[10];
	if (RECORD[rec].room >= 100){
		sprintf(dspbuf,"%02bx/%02bx/20%02bx",RECORD[rec].date,
		RECORD[rec].month,RECORD[rec].year);
		PrnString(dspbuf);
		PrnBlank(2);
		sprintf(dspbuf,"%03d",RECORD[rec].room);
		PrnString(dspbuf);
		PrnBlank(4);
		if((mode == STAin) || (mode == STAot) || (mode == STAcup)) {
			sprintf(dspbuf,"%02bx:%02bx",RECORD[rec].hour,RECORD[rec].min);
			PrnString(dspbuf);
			PrnBlank(4);
		}
		else {
			PrnString("  -  ");
			PrnBlank(4);
		}
		if(mode == STAout) {
			sprintf(dspbuf,"%02bx:%02bx",RECORD[rec].hour,RECORD[rec].min);
			PrnString(dspbuf);
			PrnBlank(4);
		}
		else {
			PrnString("  -  ");
			PrnBlank(4);
		}
		
		if (mode == STAin){
			PrnString("In\n");
		}
		else if (mode == STAot){
			PrnString("Ot\n");
		}
		else if (mode == STAcup){
			PrnString("Cup\n");
		}
		else if (out == STAin){
			PrnString("In\n");
		}
		else if (out == STAot){
			PrnString("Ot\n");
		}
		else if (out == STAcup){
			PrnString("Cup\n");
		}
		else PrnString("\n");
	}
}

void PrnBlank(char blank) {
	char blk;
	blk = blank;
	while(blk--) PrnChar(' ');       // print blank 
}

void ListRoom(void) {
	char key,slave,port;
	LcdClear();
	printf(" SLAVE PORT  ROOM\n");
	slave = 0;
	port  = 0;
	while(1) {
		WDS;
		LcdPos(20);
		printf(" [%02bd].[%02bd] - [%03d]",slave+1,port+1,ROOM[slave][port]);
		key = ScanKey();
		if(key == STOP) {
			LcdClear();
			ErrBeep(1);
			return;
		}
		else
		if(key == '1') {
			if(++slave >= NUMCARD ) slave = 0;
		}
		else
		if(key == '4') {
			if(--slave < 0) slave = NUMCARD-1;
		}
		else
		if(key == '2') {
			if(++port >= PORT) port = 0;
		}
		else
		if(key == '5') {
			if(--port < 0) port = PORT-1;
		}
	}
}

void SetTime(void) {
	unsigned char *tm;
	char key,max1,max2,dg,pos;
	LcdClear();
	printf("   SET TIME\n");
	printf("   PASSWORD [xxxx]");
	pos = 0;
	while(1) {
		WDS;
		key = ScanKey();
		if(key == STOP) {
			ErrBeep(1);
			return;
		}
		else
		if(key == ENT) {
			if(GetPassword(33)) {
				LcdClear();
				printf("    SET TIME/DATE\n");
				printf("%02bx:%02bx:%02bx %02bx/%02bx/20%02bx",RTC.hour,RTC.min,RTC.sec,RTC.date,RTC.month,RTC.year);
				pos = 0;
				LcdPos(pos+20);
				Lcdwi(0x0d);
				while(1) {
					WDS;
					switch(pos) {
						case 0:  tm = &RTC.hour;	max1='2';max2='0';dg=0; break;
						case 1:  tm = &RTC.hour;	max1='2';max2='3';dg=1; break;
						case 3:  tm = &RTC.min; 	max1='5';max2='0';dg=0; break;
						case 4:  tm = &RTC.min; 	max1='5';max2='9';dg=1; break;
						case 6:  tm = &RTC.sec; 	max1='5';max2='0';dg=0; break;
						case 7:  tm = &RTC.sec; 	max1='5';max2='9';dg=1; break;
						case 9:  tm = &RTC.date;	max1='3';max2='0';dg=0; break;
						case 10: tm = &RTC.date;	max1='3';max2='1';dg=1; break;
						case 12: tm = &RTC.month;	max1='1';max2='0';dg=0; break;
						case 13: tm = &RTC.month;	max1='1';max2='2';dg=1; break;
						case 17: tm = &RTC.year;	max1='9';max2='0';dg=0; break;
						case 18: tm = &RTC.year;	max1='9';max2='9';dg=1; break;
					}
					switch(SetDigit(tm,20+pos,max1,max2,dg)) {
						case 0: Lcdwi(0x0c); return;
						case 1: {
							if(pos < 18) pos++;
							else OverBeep(1);
							if((pos==2)||(pos==5)||(pos==8)||(pos==11)||(pos==14)) {
								pos++;
							}
							break;
						}
						case 2: {
							if(pos > 0) pos--;
							else UnderBeep(1);
							if((pos==2)||(pos==5)||(pos==8)||(pos==11)||(pos==14)) {
								pos--;
							}
							break;
						}
						case 3: {
							Lcdwi(0x0c);
							SetRTC();
							SuccBeep(1);
							return;
						}
					}
				}
			}
		}
	}
}

char SetDigit(unsigned char *digit,char pos,char max1,char max2,char dg) {
	char key,hexbuf;
	LcdPos(pos);
	while(1) {
		WDS;
		key = ScanKey();
		if(key == STOP) {
			return(0);
		}
		else if(key == DN)	return(1); 	// 1 <--  Forward		
		else if(key == UP)	return(2); 	// 2 <--  Backward		
		else if(key == ENT) return(3); 	// 4 <--  Save and Exit	

		if(isdigit(key)) {
			LcdPos(pos);
			if(!dg) {
				if(key <= max1) {
					putchar(key);
					hexbuf =   key - 0x30;
					hexbuf <<= 4;
					*digit &=  0x0f;
					*digit |=  hexbuf;
					return(1);
				}
			}
			else {
				hexbuf =   *digit;
				hexbuf &=  0xf0;
				hexbuf >>= 4;
				hexbuf +=  0x30;
				if(hexbuf  == max1) {
					if(key <= max2) {
						putchar(key);
						hexbuf	= key - 0x30;
						*digit &= 0xf0;
						*digit |= hexbuf;
						return(1);
					}
				}
				else {
					putchar(key);
					hexbuf =  key - 0x30;
					*digit &= 0xf0;
					*digit |= hexbuf;
					return(1);
				}
			}
		}
	}
}

void PrnString(char *s) {
	while(*s) PrnChar(*s++);
}

char putchar(char c) {
	if(c == '\n') {
		LCDPOS = 20;
		LcdPos(LCDPOS);
	}
	else {
		LcdPos(LCDPOS);
		Lcdwd(c);
		LCDPOS++;
	}
	return(c);
}

void LcdPos(char pos) {
	char line,col,real;
	LCDPOS = pos;
	line   = pos/20;
	col    = pos%20;
	switch(line) {
		case 0: real = 0x80+col; break;
		case 1: real = 0xc0+col; break;
	}
	Lcdwi(real);
}

void Lcdwi(char com) {
	LCDWRC = com;
	LcdWait();
}

void Lcdwd(char dat) {
	LCDWRD = dat;
	LcdWait();
}

void LcdWait(void) {
	char busy;
	do {
		busy  = (LCDRDC & 0x80);
	}while(busy);
}

void LcdSet(void) {
	Lcdwi(0x38);
	Lcdwi(0x0c);
	LcdClear();
}

void LcdClear(void) {
	LCDPOS = 0;
	Lcdwi(0x0c);
	Lcdwi(1);
}

void SendBlink(char led,char mode) {
	Tx485(':');          // Status  0 = off    
	Tx485(0xfe);		 // 		1 = slow   
	Tx485(led); 		 // 		3 = on	   
	Tx485(mode);
}

/*void Out485_16(unsigned char addr,unsigned int dat) {
	unsigned char c,d,chksum;
	LEDTX  = 0;

	c = (char)(dat & 0x00ff);
	d = (char)((dat >> 8) & 0x00ff);

	chksum = 0;
	chksum = chksum+':';
	chksum = chksum+addr;
	chksum = chksum+c;
	chksum = chksum+d;
	chksum = ~chksum;
	chksum = chksum+1;
	Tx485(':');
	Tx485(addr);
	Tx485(c);
	Tx485(d);
	Tx485(chksum);
	LEDTX  = 1;
} */

void Out485_16(unsigned char addr,unsigned int dat) {
	unsigned char out[8],chksum;
	LEDTX  = 0;

	out[0] = 0;
	out[1] = 0;
	out[2] = 0;
	out[3] = 0;
	out[4] = 0;
	out[5] = 0;
	out[6] = 0;
	out[7] = 0;

	if (dat & 0x0001) out[0] |= TSTATUS[addr][1].mode << 4;
	if (dat & 0x0002) out[0] |= TSTATUS[addr][2].mode;
	if (dat & 0x0004) out[1] |= TSTATUS[addr][3].mode << 4;
	if (dat & 0x0008) out[1] |= TSTATUS[addr][4].mode;
	if (dat & 0x0010) out[2] |= TSTATUS[addr][5].mode << 4;
	if (dat & 0x0020) out[2] |= TSTATUS[addr][6].mode;
	if (dat & 0x0040) out[3] |= TSTATUS[addr][7].mode << 4;
	if (dat & 0x0080) out[3] |= TSTATUS[addr][8].mode;
	if (dat & 0x0100) out[4] |= TSTATUS[addr][9].mode << 4;
	if (dat & 0x0200) out[4] |= TSTATUS[addr][10].mode;
	if (dat & 0x0400) out[5] |= TSTATUS[addr][11].mode << 4;
	if (dat & 0x0800) out[5] |= TSTATUS[addr][12].mode;
	if (dat & 0x1000) out[6] |= TSTATUS[addr][13].mode << 4;
	if (dat & 0x2000) out[6] |= TSTATUS[addr][14].mode;
	if (dat & 0x4000) out[7] |= TSTATUS[addr][15].mode << 4;
	if (dat & 0x8000) out[7] |= TSTATUS[addr][16].mode;

	chksum = 0;
	chksum = chksum+':';
	chksum = chksum+addr;
	chksum = chksum+out[0];
	chksum = chksum+out[1];
	chksum = chksum+out[2];
	chksum = chksum+out[3];
	chksum = chksum+out[4];
	chksum = chksum+out[5];
	chksum = chksum+out[6];
	chksum = chksum+out[7];
	chksum = ~chksum;
	chksum = chksum+1;
	Tx485(':');
	Tx485(addr);
	Tx485(out[0]);
	Tx485(out[1]);
	Tx485(out[2]);
	Tx485(out[3]);
	Tx485(out[4]);
	Tx485(out[5]);
	Tx485(out[6]);
	Tx485(out[7]);
	Tx485(chksum);
	LEDTX  = 1;
}

void Tx485(char ser) {
	while(!TI);
	TI	 = 0;
	SBUF = ser;
}

void Delay(unsigned int delaytime) {
	//TL0 = TIM0LO;
	//TH0 = TIM0HI;
	ET0 = 1;
	TR0 = 1;
	Msecflag = 0;
	while(delaytime > 0) {
		WDS;
		if(Msecflag){
			Msecflag = 0;
			delaytime -= 2;
		}
	}
	//ET0 = 0;
	//TR0 = 0;
}

void timer0() interrupt 1 using 0 {
	EA = 0;
	TL0 = TIM0LO;
	TH0 = TIM0HI;
	Msecflag = 1;
	Dspflag = 1;
	EA	= 1;
}

void ClearRam(void) {
	char  i,j;
	ClrDisbuf();
	CUPTIME 	= 15;		  // Clean up time 15 min	
	NUMCARD		= 8;		  // 8 Slave  Max		   	
	OTHOUR		= 120;		  // Over time hours		
	REMAIN1     = 30;
	REMAIN2     = 20;
	PrintRef	= 1;		  // Clear Print Reference		
	IMM_DAY     = 1;
	PASSWORD[0] = '0';
	PASSWORD[1] = '0';
	PASSWORD[2] = '0';
	PASSWORD[3] = '0';
	ClearRecord();

	for(i=0; i<NUMCARD; i++) {
		ROOMOUT[i]  = 0;
		for(j=0; j<PORT; j++) {
			ROOM[i][j] = 0;
			TSTATUS[i][j].mode = 0xff;
			TSTATUS[i][j].hour = 0;
			TSTATUS[i][j].min  = 0;
			TSTATUS[i][j].totalmin = 0;
		}
	}

	NormBeep(0);
	LcdClear();
	printf("    Clear Memory\n");
	printf("     Completed");
	Delay(1000);
	SuccBeep(0);
}

void ClearRecord(void) {
	int i;
	for(i=0;i < MAXREC;i++) {
		RECORD[i].date	= 0xff;
		RECORD[i].month = 0xff;
		RECORD[i].year	= 0xff;
		RECORD[i].room	= 0xff;
		RECORD[i].hour	= 0xff;
		RECORD[i].min	= 0xff;
		RECORD[i].mode	= 0xff;
		RECORD[i].out	= 0xff;
	}
	RecordIndex = 0;		  // Clear Period index 
	PeriodStart = 0;
	PeriodEnd   = 0;
}

void ClearRecCompleted(void) {
	NormBeep(0);
	LcdClear();
	printf("    Clear Record\n");
	printf("     Completed");
	Delay(1000);
	SuccBeep(0);
}

void ClrDisbuf(void) {
	char  i;
	for(i=0;i<sizeof(DISPBUF);i++) DISPBUF[i] = '\0';
}

char ScanKey(void) {
	char row,col,col1,key;
	unsigned char i;
	char code keytab[4][7] = {{'U','1','2','3','I','+',0},
							  {'D','4','5','6','O','-',0},
							  {'A','7','8','9','C','P',0},
                              {'B','*','0','#','E','S',0}};

	for(col=0;col<6;col++) {
		switch(col) {
			case 0: col1 = 0x3e; break;
			case 1: col1 = 0x3d; break;
			case 2: col1 = 0x3b; break;
			case 3: col1 = 0x37; break;
			case 4: col1 = 0x2f; break;
			case 5: col1 = 0x1f; break;
		}
		KEYCOL = (KEYCOL & 0xc0 | col1);
		for(i=0;i<100;i++);
		row = KEYROW & 0x0f;
		if(row != 0x0f) {
			for(i=0;i<100;i++);  	// Debounce 
			row = KEYROW & 0x0f;
			if(row != 0x0f) {
				switch(row) {
					case 0x0e: row = 0; break;
					case 0x0d: row = 1; break;
					case 0x0b: row = 2; break;
					case 0x07: row = 3; break;
				}
				key = keytab[row][col];
				if(!Keyflag) {
					Keyflag = 1;
					KeyCnt = KEYDEL1;
					NormBeep(0);
					return(key);
				}
				else {
					if (--KeyCnt) return '\0';
					else {
						SuccBeep(0);
						KeyCnt = KEYDEL2;
						return (key);
					}
				} 
			}
		}
	}
	Keyflag = 0;
	return '\0';
}

/*char ScanKey(void) {
	char code keytab[4][8] = {
	'U',0,'1','2','3','I',0,'+',
  'D',0,'4','5','6','O',0,'-',
  'A',0,'7','8','9','C',0,'P',
  'B',0,'*','0','#','E',0,'S'};

	KEYBUF = '\0';
	if (RI){
		RI = 0;
		if(!Keyflag) {
			KEYBUF	 = keytab[SBUF];
			//KeyCnt = KEYDEL1;
			Keyflag = 1;
			NormBeep(0);
		}
	}
	else {
		Keyflag = 0;
	}
	return(KEYBUF);
}*/

void Beep(char freq,int time) {
	unsigned char data i; int j;
	for(j=0;j<time;j++) {
		WDS;
		SOUNDB = 0;
		for(i=0;i<freq;i++);
		SOUNDB = 1;
		for(i=0;i<freq;i++);
	}
}

void PowerBeep(void) {	 	// Power Beep		 
	Beep(40,400);
}

void ErrBeep(bit dly) { 	// Error Beep		
	if(dly) Delay(100);
	Beep(100,300);
}


void OverBeep(bit dly) {	// Over range Beep	
	if(dly) Delay(100);
	Beep(80,100);
}

void UnderBeep(bit dly) {	// Under range Beep 
	if(dly) Delay(100);
	Beep(90,100);
}

void SuccBeep(bit dly) {	// Success Beep 	
	if(dly) Delay(100);
	Beep(40,200);
}

void NormBeep(bit dly) {	// Normal Beep		
	if(dly) Delay(100);
	Beep(30,100);
}

void PrnChar(char c) {
	char p = 0;
	int  i = 0;
    do {
		WDS;
		p = (PRNBSY & 0x80);
		if(++i > 10000) {
			LcdClear();
			printf("   PRINTER ERROR\n");
			i = 0;
		}
	}while(p);
    PRNDAT = c;
	p	   = PRNSTB;
	PRNSTB = p & 0x7f;
	PRNSTB = p | 0x80;
}

void InitRTC (void){
	unsigned char dat;
	Getreg(RTCADDR,0,&dat);						//Secunds
	dat &= 0x7f;
	Putreg(RTCADDR,0,dat);
	Getreg(RTCADDR,2,&dat);						//Hours
	dat &= 0xbf;
	Putreg(RTCADDR,2,dat);
}

void SetRTC(void) {
	Putreg(RTCADDR,0,RTC.sec);
	Putreg(RTCADDR,1,RTC.min);
	Putreg(RTCADDR,2,RTC.hour);
	Putreg(RTCADDR,3,RTC.day);
	Putreg(RTCADDR,4,RTC.date);
	Putreg(RTCADDR,5,RTC.month);
	Putreg(RTCADDR,6,RTC.year);
}

void GetRTC(void) {
	Getreg(RTCADDR,0,&RTC.sec);
	Getreg(RTCADDR,1,&RTC.min);
	Getreg(RTCADDR,2,&RTC.hour);
	Getreg(RTCADDR,3,&RTC.day);
	Getreg(RTCADDR,4,&RTC.date);
	Getreg(RTCADDR,5,&RTC.month);
	Getreg(RTCADDR,6,&RTC.year);
}

bit Putreg(char addr,char reg,char dat) {
	if((SCL==0) || (SDA==0)) {
		return(1);
	}
	Start();
	if(Txbyte(addr)) {
		Stop();
		return(1);
	}
	if(Txbyte(reg)) {
		Stop();
		return(1);
	}
	Txbyte(dat);
	Stop();
	return(0);
}

bit Getreg(char addr,char reg,char *dat) {
	if((SCL==0) || (SDA==0)) {
		return(1);
	}
	Start();
	if(Txbyte(addr)) {
		Stop();
		return(1);
	}
	if(Txbyte(reg)) {
		Stop();
		return(1);
	}
	Start();
	if(Txbyte(addr|0x01)) {
		Stop();
		return(1);
	}
	*dat = Rxbyte();
	SDA  = 1;
	Plscl();
	Stop();
	return(0);
}

bit Txbyte(unsigned char com) {
	unsigned char data loop;
	bit erflag;
	for(loop=0; loop<8; loop++) {
		SDA   = com & 0x80;
		com <<= 1;
		Plscl();
	}
	SDA = 1;
	Setscl();
	erflag = SDA;
	Clrscl();
	return(erflag);
}

char Rxbyte(void) {
	unsigned char data dat,loop;
	dat = 0;
	for(loop=0; loop<8; loop++) {
		dat <<= 1;
		Setscl();
		dat |= SDA;
		Clrscl();
	}
	return(dat);
}

void Start(void) {
	SDA = 1;
	Setscl();
	SDA = 0;
	Bitdl();
	Clrscl();
}

void Stop(void) {
	SDA = 0;Setscl();
	SDA = 1;Bitdl();
}

void Plscl(void) {
	Setscl();
	Clrscl();
}

void Setscl(void) {
	SCL = 1;
	while(!SCL);
	Bitdl();
}

void Clrscl(void) {
	SCL = 0;
	Bitdl();
}

void Bitdl(void) {
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
	_nop_();
}

/*** edit 15/7/46  ****/
void SetAlarm(char rem) {
	bit  dispfg  = 1;  
	char key;
	char remm;
		 if(rem == 0) remm = REMAIN1;
	else if(rem == 1) remm = REMAIN2;
	LcdClear();
	while(1) {
		WDS;
		if(dispfg) {
			dispfg = 0;
			LcdClear();
			printf("Set Alarm%01bd[%02bd] Min",rem+1,remm);
        }
		key = ScanKey();
		if(key == UP) {
			if(++remm > 59) remm = 0;
			dispfg = 1;
		}
	   	else if(key == DN) {
			if(--remm < 0) remm = 59;
			dispfg = 1;
		}		
		else if(key == ENT) {
        		 if(rem == 0) REMAIN1 = remm;
			else if(rem == 1) REMAIN2 = remm;
   			SuccBeep(1);
			return;
		}
	}
}

void SetOThour(void) {
	bit  dispfg;
	char key;
	int  hour_mem;
	LcdClear();
	printf("SET OT Hour[    ]H\n");
	printf("Press key 1 - 9");
	dispfg = 1;
	hour_mem = OTHOUR;
	while(1) {
		WDS;
		if(dispfg) {
			dispfg = 0;
			LcdPos(12);
			switch(hour_mem) {
				case 60:   printf("1:00"); break;
				case 90:   printf("1:30"); break;						
				case 105:  printf("1:45"); break;					
				case 120:  printf("2:00"); break;
				case 150:  printf("2:30"); break;
				case 165:  printf("2:45"); break;						
				case 180:  printf("3:00"); break;
				case 210:  printf("3:30"); break;						
				case 225:  printf("3:45"); break;
			}
		}
		key = ScanKey();
		if(('1' <= key) && (key <= '9')) {
			dispfg = 1;
			switch(key) {
				case '1':  hour_mem = 60;  break;	// 1:00
   				case '2':  hour_mem = 90;  break;   // 1:30
				case '3':  hour_mem = 105; break;   // 1:45
				case '4':  hour_mem = 120; break;   // 2:00
				case '5':  hour_mem = 150; break;   // 2:30
				case '6':  hour_mem = 165; break;   // 2:45
				case '7':  hour_mem = 180; break;   // 3:00
				case '8':  hour_mem = 210; break;   // 3:30
                case '9':  hour_mem = 225; break;   // 3:45
            }
		}
		else
		if(key == ENT) {
			OTHOUR = hour_mem;	
			SuccBeep(1);
			return;
		}
		else
		if(key == STOP) {
			ErrBeep(1);
			return;
		}
	}
}

/*--- 10-5-47  Direct print --------------*/
void Direct_Prn(char day) {
	bit leapYear;
	char  startDate,endDate,startMonth,endMonth,i;	
	
	if ((RTC.year % 4) == 0) leapYear = 1;
	else leapYear = 0;
	endDate  = RTC.date;
	endMonth = RTC.month;	
	
	if((endDate - day) > 0) {
		startDate  = (endDate - day);
        startMonth = endMonth;
	}
	else if((endDate - day) == 0) {
		startDate  = endDate;
        startMonth = endMonth;
    }
	else if((endDate - day) < 0) {
		i = day - 1;
		switch(endMonth) {
			case 1:		startDate = endDate;  startMonth = 1;  break;	
			case 2: 	startDate = (31-i);   startMonth = 1;  break;	
			case 3: 	if (leapYear) {
						startDate = (29-i);   startMonth = 2;  break;	
						} else
						startDate = (28-i);   startMonth = 2;  break;
			case 4: 	startDate = (31-i);   startMonth = 3;  break;	
			case 5: 	startDate = (30-i);   startMonth = 4;  break;	
			case 6: 	startDate = (31-i);   startMonth = 5;  break;	
			case 7: 	startDate = (30-i);   startMonth = 6;  break;	
			case 8: 	startDate = (31-i);   startMonth = 7;  break;	
			case 9: 	startDate = (31-i);   startMonth = 8;  break;	
			case 10:	startDate = (30-i);   startMonth = 9;  break;	
			case 11:	startDate = (31-i);   startMonth = 10; break;	
			case 12:	startDate = (30-i);   startMonth = 11; break;	
		}
	}
	PrintSum(startDate,endDate,startMonth,endMonth);
}

void Set_Imm_Prn(void) {
	bit  dispfg;
	char key;
	int  day_mem;
	LcdClear();
	printf("Set Immediately Prn\n");
	printf("Press key 1 - 3");
	dispfg = 1;
	day_mem = IMM_DAY;
	while(1) {	
		WDS;
		key = ScanKey();
		if(('1' <= key) && (key <= '3')) {
			dispfg = 1;
			switch(key) {
				case '1': day_mem = 1; break;	// 1:00
   				case '2': day_mem = 2; break;   // 1:30
				case '3': day_mem = 3; break;   // 1:45
	        }
		}
		else
		if(key == ENT) {
			IMM_DAY = day_mem;	
			SuccBeep(1);
			return;
		}
		else
		if(key == STOP) {
			ErrBeep(1);
			return;
		}
	}
}
