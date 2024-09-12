import 'package:pocket_functions/pocket_functions.dart';

void calculate(PocketRequest request) {
  var result = {"add": 6 * 7};

  request.response
      .addHeader("content-type", "application/json")
      .write(result)
      .close();
}
