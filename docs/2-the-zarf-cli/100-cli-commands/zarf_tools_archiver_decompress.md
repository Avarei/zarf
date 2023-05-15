# zarf tools archiver decompress
<!-- Auto-generated by docs/gen-cli-docs.sh -->

Decompress an archive or Zarf package based off of the source file extension.

```
zarf tools archiver decompress {ARCHIVE} {DESTINATION} [flags]
```

## Options

```
      --decompress-all   Decompress all layers in the archive
  -h, --help             help for decompress
```

## Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

## SEE ALSO

* [zarf tools archiver](zarf_tools_archiver.md)	 - Compress/Decompress generic archives, including Zarf packages.