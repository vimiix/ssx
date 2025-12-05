# Installation

## Download Online

You can download the package for your platform from the GitHub [release page](https://github.com/vimiix/ssx/releases).

> Download link: [https://github.com/vimiix/ssx/releases](https://github.com/vimiix/ssx/releases)

After downloading, extract the archive to get the `ssx` binary file (`ssx.exe` on Windows).

For example, on `linux x86_64`:

```bash
tar -xvf ssx_vX.Y.Z_linux_x86_64.tar.gz
```

Place the extracted `ssx` binary in any directory included in your `$PATH`. You can also use it directly via `./ssx` or with an absolute path. For convenience, it's recommended to place it in a directory included in `$PATH`, such as `/usr/local/bin`. If the directory containing ssx is not in `$PATH`, you can add it:

```bash
echo 'export PATH=<ssx_directory>:$PATH' >> ~/.bashrc
source ~/.bashrc
```

Now you can use `ssx` directly from any terminal, in any directory.

## Build from Source

If the release page doesn't provide a package for your platform, you can compile it yourself from source:

> Local compilation requires Go 1.19+

```bash
git clone https://github.com/vimiix/ssx.git
cd ssx
make ssx
```

After successful compilation, the ssx binary will be generated in the **dist/** directory.
