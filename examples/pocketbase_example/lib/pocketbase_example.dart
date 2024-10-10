import 'package:pocket_functions/pocket_functions.dart';
import 'package:pocket_functions/entry_point.dart';
import 'package:pocketbase/pocketbase.dart';

final pb = PocketBase('http://192.168.0.17:8090');

main() async {
  entryPoint.onRequest((request) async {
    if (request.httpMethod == "GET") {
      final result = await pb.collection("example").getList(
            page: int.tryParse(request.params['page']?[0] ?? "1") ?? 1,
            perPage: 20,
            filter: 'status = true && created >= "2022-08-01"',
            sort: '-created',
          );

      var response =
          request.response.addHeader("content-type", "application/json");

      var items = result.items.map((e) {
        return e.toJson();
      }).toList();
      response.write(items);
      response.close();
    } else if (request.httpMethod == "POST") {
      var name = request.jsonBody()!["name"];
      var status = request.jsonBody()!["status"];

      final body = <String, dynamic>{"name": name, "status": status};

      final record = await pb.collection('example').create(body: body);

      var response =
          request.response.addHeader("content-type", "application/json");
      response.write([record.toJson()]);
      response.close();
    }
  });
}
