m2LocalRepoTools
================

通过 HTTP API 将本地 Maven 仓库的制品并行快速上传到 Nexus 私服的工具。

# 0 功能特性
* 通过 RESTful API **推送本地制品**到 Nexus 私服（无需安装 Maven/JDK）
* 自动识别 **releases / snapshots** 版本，分别上传到对应的仓库
* 支持**清理**本地仓库中指定后缀的文件
* 支持**并行上传**，可通过 worker 数量配置并发度
* 支持 YAML 配置文件或命令行参数两种配置方式

# 1 构建
```bash
# 编译
go build -ldflags="-s -w" -v -o m2LocalRepoTools ./main.go

# 打包（生成 m2LocalRepoTools-bin.zip，包含多架构多系统二进制）
./build.sh
```

## 打包结构
```
m2LocalRepoTools
├── config.yaml.template    # 配置文件模板
├── README.md               # 英文文档
├── README_CN.md            # 中文文档
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

# 2 支持的制品类型
| 打包类型           | 分类器             | 示例                                                    |
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

上传时自动忽略以下后缀的文件：
`.DS_Store`、`.asc`、`.lastUpdated`、`.md5`、`.repositories`、`.sha1`、`.sha256`、`.sha512`、`.xml`

# 3 快速开始

## 3.1 通过命令行参数
```bash
chmod +x m2LocalRepoTools

# 查看帮助信息
./m2LocalRepoTools -help

# 推送制品到私服
./m2LocalRepoTools [OPTIONS] push

# 清理本地仓库
./m2LocalRepoTools [OPTIONS] clean
```

其中 `[OPTIONS]` 说明

| 参数              | 默认值                      | 说明                                                    |
|-----------------|--------------------------|-------------------------------------------------------|
| `-help`         |                          | 显示帮助信息                                                |
| `-repo`         | `${HOME}/.m2/repository` | 本地仓库路径                                                |
| `-releaseUrl`   | -                        | Nexus releases 仓库地址，仅上传 release 版本（三者至少配置一个）          |
| `-snapshotsUrl` | -                        | Nexus snapshots 仓库地址，仅上传 snapshot 版本（三者至少配置一个）        |
| `-mixedUrl`     | -                        | Nexus mixed 仓库地址，同时支持 release 和 snapshot 版本（三者至少配置一个） |
| `-username`     | -                        | Nexus 用户名                                             |
| `-password`     | -                        | Nexus 密码                                              |
| `-worker`       | 4                        | 并行上传线程数                                               |
| `-verbose`      |                          | 开启调试输出                                                |

版本路由规则：
* release 版本优先上传到 `releaseUrl`，若未配置则使用 `mixedUrl`
* snapshot 版本（版本号以 `-SNAPSHOT` 结尾）优先上传到 `snapshotsUrl`，若未配置则使用 `mixedUrl`
* 若对应版本类型无可用 URL，该制品将被跳过

一个上传本地repository指定目录包的示例
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

## 3.2 通过 YAML 配置文件
```bash
# 推送本地 Maven 仓库的制品到私服
./m2LocalRepoTools --config ./config.yaml push

# 清理本地 Maven 仓库中配置列表指定的文件
./m2LocalRepoTools --config ./config.yaml clean
```

完整配置请参考 [`config.yaml.template`](static/config.yaml.template)：
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
