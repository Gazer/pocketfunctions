import 'dart:io';
import 'package:archive/archive.dart';
import 'package:path/path.dart' as p;

Future<void> createPackageZip(String packagePath, String outputPath) async {
  final zipEncoder = ZipEncoder();
  final archive = Archive();

  void addFileToArchive(String filePath, String relativePath) {
    final file = File(filePath);
    final fileBytes = file.readAsBytesSync();
    final archiveFile = ArchiveFile(relativePath, fileBytes.length, fileBytes);
    archive.addFile(archiveFile);
  }

  final pubspecFile = File(p.join(packagePath, 'pubspec.yaml'));
  if (pubspecFile.existsSync()) {
    addFileToArchive(pubspecFile.path, 'pubspec.yaml');
  } else {
    print('No pubspec.yaml');
  }

  final libDir = Directory(p.join(packagePath, 'lib'));
  if (libDir.existsSync()) {
    libDir.listSync(recursive: true).forEach((file) {
      if (file is File) {
        final relativePath = p.relative(file.path, from: packagePath);
        addFileToArchive(file.path, relativePath);
      }
    });
  } else {
    print('No lib/');
  }

  final zipFile = File(outputPath);
  zipFile.writeAsBytesSync(zipEncoder.encode(archive)!);
  print('ZIP: $outputPath');
}
