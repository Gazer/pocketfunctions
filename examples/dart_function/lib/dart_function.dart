import 'package:pocket_functions/pocket_functions.dart';

@EntryPoint()
Future<void> calculate(PocketRequest request) async {
  var a = int.tryParse(request.params["a"]?[0] ?? "") ?? 0;
  var b = int.tryParse(request.params["b"]?[0] ?? "") ?? 0;
  var result = {"add": a + b};

  request.response
      .addHeader("content-type", "application/json")
      .write(result)
      .close();
}
