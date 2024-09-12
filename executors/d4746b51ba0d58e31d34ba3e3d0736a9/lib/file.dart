import 'package:pocket_functions/pocket_functions.dart';
import 'package:pocketbase/pocketbase.dart';

final pb = PocketBase('http://127.0.0.1:8090');

Future<void> calculate(PocketRequest request) async {
  final result = await pb.collection('example').getList(
        page: 1,
        perPage: 20,
        filter: 'status = true && created >= "2022-08-01"',
        sort: '-created',
      );

  var response = request.response.addHeader("content-type", "application/json");

  var items = result.items.map((e) {
    return e.toJson();
  }).toList();
  response.write(items);
  response.close();
}
