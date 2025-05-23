import 'package:pocket_functions/pocket_functions.dart';

void main() {
  entryPoint.onRequest((request) {
    var a = int.tryParse(request.params["a"]?[0] ?? "") ?? 0;
    var b = int.tryParse(request.params["b"]?[0] ?? "") ?? 0;
    var result = {
      "add": a + b,
      "minus:": a - b,
      "mult": a * b,
      "div": a / b,
      "other": a ^ b
    };

    request.response
        .addHeader("content-type", "application/json")
        .write(result)
        .close();
  });
}
