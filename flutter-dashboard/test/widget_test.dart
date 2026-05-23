import 'package:flutter_test/flutter_test.dart';

import 'package:flutter_dashboard/src/app.dart';

void main() {
  testWidgets('Dashboard app builds', (WidgetTester tester) async {
    await tester.pumpWidget(const DashboardApp());

    expect(find.text('Overview'), findsOneWidget);
  });
}
