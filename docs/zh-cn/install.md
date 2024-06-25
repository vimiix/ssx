# 安装

## 在线下载

你可以通过 github 的 [release 页面](https://github.com/vimiix/ssx/releases)下载对应平台的软件包

> 下载地址：[https://github.com/vimiix/ssx/releases](https://github.com/vimiix/ssx/releases)

下载后解压压缩包即可得到 `ssx` 二进制文件（windows 平台为 `ssx.exe`）

以 `linux x86_64` 为例:

```bash
tar -xvf ssx_vX.Y.Z_linux_x86_64.tar.gz
```

将解压得到的 `ssx` 二进制文件存放到任意 `$PATH` 中存在的目录中即可，当然也可以直接通过 `./ssx` 或绝对路径的方式使用，为了便于使用，建议还是放到 `$PATH` 中包含的目录中，比如 `/usr/local/bin`，如果 ssx 所在的目录不在 `$PATH` 环境变量中，可以通过配置添加：

```bash
echo 'export PATH=<ssx所在目录>:$PATH' >> ~/.bashrc
source ~/.bashrc
```

此时，我们就可以任何时候打开终端，在任意目录下直接使用 `ssx` 了。

## 源代码安装

如果 release 页面没有提供你所使用的平台的包，你可以可以通过源码来自己编译出对应平台的包：

> 本地编译需要安装 go 1.19+

```bash
git clone https://github.com/vimiix/ssx.git
cd ssx
make ssx
```

编译成功后，会在 **dist/** 目录下生成 ssx 的二进制文件。
