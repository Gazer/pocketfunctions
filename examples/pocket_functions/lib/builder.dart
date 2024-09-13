import 'package:build/build.dart';
import 'package:source_gen/source_gen.dart';
import 'entry_point_generator.dart';

Builder entryPointBuilder(BuilderOptions options) => LibraryBuilder(
      EntryPointGenerator(),
      generatedExtension: '.entrypoint.g.dart',
    );
