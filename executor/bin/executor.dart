import 'package:executor/file.dart';
import 'package:pocket_functions/pocket_functions.dart';

main(List<String> arguments) async {
  var request = PocketRequest();
  calculate(request);
  print(request.response);
}
