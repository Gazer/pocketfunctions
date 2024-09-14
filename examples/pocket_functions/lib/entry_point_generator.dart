import 'package:build/build.dart';
import 'package:source_gen/source_gen.dart';
import 'package:analyzer/dart/element/element.dart';
import 'package:pocket_functions/entry_point.dart';

class EntryPointGenerator extends GeneratorForAnnotation<EntryPoint> {
  @override
  String generateForAnnotatedElement(
    Element element,
    ConstantReader annotation,
    BuildStep buildStep,
  ) {
    if (element is FunctionElement) {
      final functionName = element.name;
      return '''
      import 'package:pocket_functions/pocket_functions.dart';
      import '${element.librarySource.uri}';

      Future<void> pocketFunctionsEntryPoint(PocketRequest request) async {
        await $functionName(request);
      }
      ''';
    }
    return '';
  }
}
