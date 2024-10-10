import 'package:pocket_functions/pocket_functions.dart';
import 'package:pocket_functions/entry_point.dart';

main() {
  entryPoint.cron("*/1 * * * *", () {
    print("${DateTime.now()} something");
  });
}
