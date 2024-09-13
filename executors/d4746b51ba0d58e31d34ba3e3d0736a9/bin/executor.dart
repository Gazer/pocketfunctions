import 'dart:io';

import 'package:executor/file.dart';
import 'package:pocket_functions/pocket_functions.dart';

main(List<String> arguments) async {
  var request = PocketRequest.fromEnvironment(Platform.environment);
  await calculate(request);
  print(request.response);
}
