package resourcemanager

import (
  "archive/zip"
  "archive/tar"
  "compress/gzip"
  "fmt"
  "os"
  "os/exec"
  "bufio"
  "strings"
  "bytes"
  "io"
  "path/filepath"
  "github.com/pkg/errors"
  util "server/service/tq-service/util"
)

type FileManager struct{
  file_dir string //

}

func (fm *FileManager) Init(){
  util.Info("FileManager init.......")
}

// 写文件
func (fm *FileManager) WriteMaptoFile(m []string, filePath string) error {
  f, err := os.Create(filePath)
  if err != nil {
    fmt.Printf("create map file error: %v\n", err)
    return err
  }
  defer f.Close()

  w := bufio.NewWriter(f)
  for _, v := range m {
    fmt.Fprintln(w, v)
  }

  return w.Flush()
}

//拷贝
func (fm *FileManager) CopyFile(dstName, srcName string) (written int64, err error) {
    src, err := os.Open(srcName)
    if err != nil {
        return
    }
    defer src.Close()
    dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        return
    }
    defer dst.Close()
    return io.Copy(dst, src)
}

//is zip
func (fm *FileManager) IsZip(zipPath string) bool {
    f, err := os.Open(zipPath)
    if err != nil {
        return false
    }
    defer f.Close()

    buf := make([]byte, 4)
    if n, err := f.Read(buf); err != nil || n < 4 {
        return false
    }

    return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// 文件unzip
func (fm *FileManager) UnZip(archive, target string) error {
    reader, err := zip.OpenReader(archive)
    if err != nil {
        return err
    }

    if err := os.MkdirAll(target, 0755); err != nil {
        return err
    }

    for _, file := range reader.File {
        path := filepath.Join(target, file.Name)
        if file.FileInfo().IsDir() {
            os.MkdirAll(path, file.Mode())
            continue
        }

        fileReader, err := file.Open()
        if err != nil {
            return err
        }
        defer fileReader.Close()

        targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
        if err != nil {
            return err
        }
        defer targetFile.Close()

        if _, err := io.Copy(targetFile, fileReader); err != nil {
            return err
        }
    }

    return nil
}

//压缩为zip格式
//source为要压缩的文件或文件夹, 绝对路径和相对路径都可以
//target是目标文件
//filter是过滤正则(Golang 的 包 path.Match)
func (fm *FileManager) Zipit(source, target, filter string) error {
	var err error
	if isAbs := filepath.IsAbs(source); !isAbs {
		source, err = filepath.Abs(source) // 将传入路径直接转化为绝对路径
		if err != nil {
			return errors.WithStack(err)
		}
	}
	//创建zip包文件
	zipfile, err := os.Create(target)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := zipfile.Close(); err != nil{
			  util.Infof("*File close error: %s, file: %s", err.Error(), zipfile.Name())
		}
	}()

	//创建zip.Writer
	zw := zip.NewWriter(zipfile)

	defer func() {
		if err := zw.Close(); err != nil{
			  util.Infof("zipwriter close error: %s", err.Error())
		}
	}()

	info, err := os.Stat(source)
	if err != nil {
		return errors.WithStack(err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return errors.WithStack(err)
		}

		//将遍历到的路径与pattern进行匹配
		ism, err := filepath.Match(filter, info.Name())

		if err != nil {
			return errors.WithStack(err)
		}
		//如果匹配就忽略
		if ism {
			return nil
		}
		//创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.WithStack(err)
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		//写入文件头信息
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return errors.WithStack(err)
		}

		if info.IsDir() {
			return nil
		}
		//写入文件内容
		file, err := os.Open(path)
		if err != nil {
			return errors.WithStack(err)
		}

		defer func() {
			if err := file.Close(); err != nil{
        util.Infof("*File close error: %s, file: %s", err.Error(), file.Name())
			}
		}()
		_, err = io.Copy(writer, file)

		return errors.WithStack(err)
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// 获取行数
func (fm *FileManager) GetlineCnt(fileName string) int{

	file,err := os.Open(fileName)
	if err != nil{
		return 0
	}
	defer file.Close()
	fd:=bufio.NewReader(file)
	count :=0
	for {
		_,err := fd.ReadString('\n')
		if err!= nil{
			break
		}
		count++
	}
  return count
}

// Tar
func (fm *FileManager) Tar(src string , dst string) (err error) {
    // 创建文件
    fw, err := os.Create(dst)
    if err != nil {
        return
    }
    defer fw.Close()

    // 将 tar 包使用 gzip 压缩，其实添加压缩功能很简单，
    // 只需要在 fw 和 tw 之前加上一层压缩就行了，和 Linux 的管道的感觉类似
    gw := gzip.NewWriter(fw)
    defer gw.Close()

    // 创建 Tar.Writer 结构
    tw := tar.NewWriter(gw)
    // 如果需要启用 gzip 将上面代码注释，换成下面的
    defer tw.Close()

    // 下面就该开始处理数据了，这里的思路就是递归处理目录及目录下的所有文件和目录
    // 这里可以自己写个递归来处理，不过 Golang 提供了 filepath.Walk 函数，可以很方便的做这个事情
    // 直接将这个函数的处理结果返回就行，需要传给它一个源文件或目录，它就可以自己去处理
    // 我们就只需要去实现我们自己的 打包逻辑即可，不需要再去路径相关的事情
    return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
        // 因为这个闭包会返回个 error ，所以先要处理一下这个
        if err != nil {
            return err
        }

        // 这里就不需要我们自己再 os.Stat 了，它已经做好了，我们直接使用 fi 即可
        hdr, err := tar.FileInfoHeader(fi, "")
        if err != nil {
            return err
        }
        // 这里需要处理下 hdr 中的 Name，因为默认文件的名字是不带路径的，
        // 打包之后所有文件就会堆在一起，这样就破坏了原本的目录结果
        // 例如： 将原本 hdr.Name 的 syslog 替换程 log/syslog
        // 这个其实也很简单，回调函数的 fileName 字段给我们返回来的就是完整路径的 log/syslog
        // strings.TrimPrefix 将 fileName 的最左侧的 / 去掉，
        // 熟悉 Linux 的都知道为什么要去掉这个
        hdr.Name = strings.TrimPrefix(fileName, string(filepath.Separator))

        // 写入文件信息
        if err := tw.WriteHeader(hdr); err != nil {
            return err
        }

        // 判断下文件是否是标准文件，如果不是就不处理了，
        // 如： 目录，这里就只记录了文件信息，不会执行下面的 copy
        if !fi.Mode().IsRegular() {
            return nil
        }

        // 打开文件
        fr, err := os.Open(fileName)
        defer fr.Close()
        if err != nil {
            return err
        }

        // copy 文件数据到 tw
        n, err := io.Copy(tw, fr)
        if err != nil {
            return err
        }

        // 记录下过程，这个可以不记录，这个看需要，这样可以看到打包的过程
        util.Infof("成功打包 %s ，共写入了 %d 字节的数据\n", fileName, n)

        return nil
    })
}

func (fm *FileManager) UnTar(dst, src string) error{
    command := "tar -zxvf " + src + " -C " + dst
    cmd := exec.Command("/bin/sh", "-c", command)
    cmdErr := cmd.Run()
    return cmdErr
}

func (fm *FileManager) UnTarold(dst, src string) (err error) {
    // 打开准备解压的 tar 包
    fr, err := os.Open(src)
    if err != nil {
        return
    }
    defer fr.Close()

    // 将打开的文件先解压
    gr, err := gzip.NewReader(fr)
    if err != nil {
        return
    }
    defer gr.Close()

    // 通过 gr 创建 tar.Reader
    tr := tar.NewReader(gr)

    // 现在已经获得了 tar.Reader 结构了，只需要循环里面的数据写入文件就可以了
    for {
        hdr, err := tr.Next()

        switch {
        case err == io.EOF:
            return nil
        case err != nil:
            return err
        case hdr == nil:
            continue
        }

        // 处理下保存路径，将要保存的目录加上 header 中的 Name
        // 这个变量保存的有可能是目录，有可能是文件，所以就叫 FileDir 了……
        dstFileDir := filepath.Join(dst, hdr.Name)

        // 根据 header 的 Typeflag 字段，判断文件的类型
        switch hdr.Typeflag {
        case tar.TypeDir: // 如果是目录时候，创建目录
            // 判断下目录是否存在，不存在就创建
            if b := ExistDir(dstFileDir); !b {
                // 使用 MkdirAll 不使用 Mkdir ，就类似 Linux 终端下的 mkdir -p，
                // 可以递归创建每一级目录
                if err := os.MkdirAll(dstFileDir, 0775); err != nil {
                    return err
                }
            }
        case tar.TypeReg: // 如果是文件就写入到磁盘
            // 创建一个可以读写的文件，权限就使用 header 中记录的权限
            // 因为操作系统的 FileMode 是 int32 类型的，hdr 中的是 int64，所以转换下
            file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, 0755)
            if err != nil {
                return err
            }
            n, err := io.Copy(file, tr)
            if err != nil {
                return err
            }
            // 将解压结果输出显示
            fmt.Printf("成功解压： %s , 共处理了 %d 个字符\n", dstFileDir, n)

            // 不要忘记关闭打开的文件，因为它是在 for 循环中，不能使用 defer
            // 如果想使用 defer 就放在一个单独的函数中
            file.Close()
        }
    }

    return nil
}

// 判断目录是否存在
func ExistDir(dirname string) bool {
    fi, err := os.Stat(dirname)
    return (err == nil || os.IsExist(err)) && fi.IsDir()
}

func (fm *FileManager) Copy(dst, src string) error{
    command := "cp -rf  " + src + " " + dst
    cmd := exec.Command("/bin/sh", "-c", command)
    cmdErr := cmd.Run()
    return cmdErr
}
