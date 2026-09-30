m2LocalRepoTools
================

[中文文档](README_CN.md)

A tool to push local Maven repository artifacts to a Nexus private server via HTTP API with parallel uploading support.

# 0 Features
* **Push local artifacts** to Nexus repository via RESTful API (no Maven/JDK required)
* Auto-detect **releases / snapshots** versions and upload to the corresponding repository
* **Clean files** with specified suffixes in local repository
* **Parallel uploading** with configurable worker pool
* Configure via YAML file or CLI parameters

# 1 Build
```bash
# build
go build -ldflags="-s -w" -v -o m2LocalRepoTools ./main.go

# package (generates m2LocalRepoTools-bin.zip with multi-arch multi-os binaries)
./build.sh
```

## Package Structure
```
m2LocalRepoTools
├── config.yaml.template    # configuration template
├── README.md               # English documentation
├── README_CN.md            # Chinese documentation
├── darwin-amd64/           # macOS (Intel)
│   └── m2LocalRepoTools
├── darwin-arm64/           # macOS (Apple Silicon)
│   └── m2LocalRepoTools
├── linux-amd64/            # Linux (x86_64)
│   └── m2LocalRepoTools
├── linux-arm64/            # Linux (ARM64)
│   └── m2LocalRepoTools
├── linux-arm/              # Linux (ARM)
│   └── m2LocalRepoTools
├── windows-amd64/          # Windows (x86_64)
│   └── m2LocalRepoTools.exe
└── windows-arm64/          # Windows (ARM64)
    └── m2LocalRepoTools.exe
```

# 2 Supported Artifact Types
| packaging      | classifier      | example                                               |
|----------------|-----------------|-------------------------------------------------------|
| `${packaging}` | `${classifier}` | `${artifactId}-${version}-${classifier}.${packaging}` |
| pom            | -               | airbase-190.pom                                       |
| - / jar        | -               | slice-2.3.jar                                         |
| jar            | executable      | trino-verifier-463-executable.jar                     |
| jar            | sources         | trino-verifier-463-sources.jar                        |
| jar            | tests           | trino-verifier-463-tests.jar                          |
| jar            | test-sources    | trino-verifier-463-test-sources.jar                   |
| uexe           | -               | credValidator-2.3.0.uexe                              |
| exe            | linux-x86_64    | protoc-3.21.7-linux-x86_64.exe                        |

The following file suffixes are ignored during upload:
`.DS_Store`, `.asc`, `.lastUpdated`, `.md5`, `.repositories`, `.sha1`, `.sha256`, `.sha512`, `.xml`

# 3 Getting Started

## 3.1 Via CLI parameters
```bash
chmod +x m2LocalRepoTools

# display help information
./m2LocalRepoTools -help

# push artifacts to Nexus
./m2LocalRepoTools [OPTIONS] push

# clean local repository
./m2LocalRepoTools [OPTIONS] clean
```

Where `[OPTIONS]` are:

| Parameter       | Default value            | Description                                                                                         |
|-----------------|--------------------------|-----------------------------------------------------------------------------------------------------|
| `-help`         | -                        | Display help information                                                                            |
| `-repo`         | `${HOME}/.m2/repository` | Local repository path                                                                               |
| `-releaseUrl`   | -                        | Nexus releases repository URL, for release versions only (at least one of three required)           |
| `-snapshotsUrl` | -                        | Nexus snapshots repository URL, for snapshot versions only (at least one of three required)         |
| `-mixedUrl`     | -                        | Nexus mixed repository URL, for both release and snapshot versions (at least one of three required) |
| `-username`     | -                        | Nexus username                                                                                      |
| `-password`     | -                        | Nexus password                                                                                      |
| `-worker`       | 4                        | Number of parallel upload workers                                                                   |
| `-verbose`      | -                        | Enable debug output                                                                                 |

Version routing rules:
* Release versions are uploaded to `releaseUrl` first; falls back to `mixedUrl` if not configured
* Snapshot versions (ending with `-SNAPSHOT`) are uploaded to `snapshotsUrl` first; falls back to `mixedUrl` if not configured
* If no suitable URL is available for the version type, the artifact will be skipped

An example of pushing artifacts from a specified local repository directory:
```bash
./m2LocalRepoTools \
  -repo ~/.m2/repository/org/antlr/ \
  -releaseUrl http://localhost:8081/repository/maven-releases/ \
  -snapshotsUrl http://localhost:8081/repository/maven-snapshots/ \
  -username admin \
  -password admin123 \
  -worker 2 \
  -verbose \
  push
```

## 3.2 Via YAML config file
```bash
# push local Maven repository artifacts to Nexus
./m2LocalRepoTools --config ./config.yaml push

# clean files matching the configured suffix list in local Maven repository
./m2LocalRepoTools --config ./config.yaml clean
```

See [`config.yaml.template`](static/config.yaml.template) for full configuration:
```yaml
verbose: false
logLevel: "INFO"
localRepository: "~/.m2/repository"

push:
  releaseUrl: "http://localhost:8081/repository/maven-releases"
  snapshotsUrl: "http://localhost:8081/repository/maven-snapshots"
  # mixedUrl: "http://localhost:8081/repository/maven-mixed"
  username: "admin"
  password: "admin123"
  worker: 4
  excludeSuffixs:
    - ".DS_Store"
    - ".asc"
    - ".lastUpdated"
    - ".md5"
    - ".repositories"
    - ".sha1"
    - ".sha256"
    - ".sha512"
    - ".xml"

clean:
  suffixs:
    - ".DS_Store"
    - ".lastUpdated"
  dirRegex:
    - "${*.version}"
```
