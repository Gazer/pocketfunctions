import 'package:pocket_functions/pocket_functions.dart';

void calculate(PocketRequest request) {
  var result = {"add": 6 * 7};

  request.response
      .addHeader("content-type", "text/plain")
      .write("Hola Mundo")
      .close();
}
