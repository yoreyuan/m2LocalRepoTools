#!/usr/bin/env bash

export PATH=$PATH:/usr/local/go/bin
export CGO_ENABLED=0

OS_ARR=(
"linux"
"windows"
"darwin"
)

ARCH_ARR=(
"amd64"
"arm64"
"arm"
)

dirName="m2LocalRepoTools"
if [ -d "${dirName}" ]; then
  rm -rf ${dirName}
fi
mkdir ${dirName}
if [ -f "${dirName}-bin.zip" ]; then
  rm ${dirName}-bin.zip
fi

for osName in ${OS_ARR[@]}; do
  for archName in ${ARCH_ARR[@]}; do
    # skip unsupported combinations
    if [ "${osName}" = "windows" ] && [ "${archName}" = "arm" ]; then
      continue
    fi
    if [ "${osName}" = "darwin" ] && [ "${archName}" = "arm" ]; then
      continue
    fi

    ext=""
    if [ "${osName}" = "windows" ]; then
      ext=".exe"
    fi

    binName="m2LocalRepoTools${ext}"
    targetDir="${dirName}/${osName}-${archName}"
    mkdir -p ${targetDir}

    echo "Building ${osName}/${archName}..."
    GOOS=${osName} GOARCH=${archName} go build -ldflags="-s -w" -v -o ${targetDir}/${binName} ./main.go
    if [ $? -ne 0 ]; then
      echo "Failed to build ${osName}/${archName}"
      rm -rf ${targetDir}
      continue
    fi
    chmod +x ${targetDir}/${binName}
  done
done

cp README.md README_CN.md ${dirName}/
cp static/config.yaml.template ${dirName}/

zip -r ${dirName}-bin.zip ${dirName}/
rm -rf ${dirName}

echo "Build complete: ${dirName}-bin.zip"
