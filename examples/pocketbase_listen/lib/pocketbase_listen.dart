import 'package:pocket_functions/pocket_functions.dart';
import 'package:pocket_functions/entry_point.dart' show entryPoint;
import 'package:pocketbase/pocketbase.dart';

final pb = PocketBase('http://192.168.0.17:8090');

main() async {
  entryPoint.listen(() {
    pb.collection('example').subscribe(
      '*',
      (e) {
        print(e.action);
        print(e.record);
      },
    );
  });
}
