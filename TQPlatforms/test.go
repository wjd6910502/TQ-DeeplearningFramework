package main

import (
  //"github.com/pkg/errors"
  "strings"
  "time"
  "fmt"
  "bufio"
  "os/exec"
  "os"
  "io"
  "archive/zip"
  "archive/tar"
  "compress/gzip"
  "path/filepath"
  "github.com/pkg/errors"
  "log"
  "bytes"
  env "server/service/tq-service/env"
  util "server/service/tq-service/util"
  httprequest "server/service/tq-service/httprequest"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

  )

func main() {
   util.Initialize()
   env.Init()
   test_getcosdata()
   //test_s3upload()
  }

func test_s3upload(){
  s3 := S3Manager{}
  s3.Init()
  s3.Upload(env.S3_BUCKET,"57640ece-c678-11ea-a45d-7af6424c9416.tar.gz")

  s3.Download(env.S3_BUCKET,"57640ece-c678-11ea-a45d-7af6424c9416.tar.gz")
}



func test_getcosdata(){
  cosid := "57640ece-c678-11ea-a45d-7af6424c9416"
  tarname := GetCosData(cosid)
  print("tarname =",tarname)
}

func test_slicedata(){
  cosid := "57640ece-c678-11ea-a45d-7af6424c9416"
  tarname := GetCosData(cosid)
  print("tarname =",tarname)
  /*
   os.Mkdir(cosid,0755)
   err := UnTar(cosid,tarname)
   if err != nil {
      print("err = ",err.Error())
  }*/

  SliceData2S3(3,cosid,tarname)
}

func test_shell() {
  //shellPath := fmt.Sprintf("./test.sh %d %s %s %s %s %s %s %s %v %v",4,"1","11","42","0021","/tq","tq","1111111",REPORT_adr,REPORT_tkn)
  //cmd := exec.Command("/bin/bash","-c",shellPath) //Cmd init
  //fmt.Println("cmd = %v",cmd.Args)
  //util.Infof("Start Exec Cmd Arg: %v , Process PID: %v ",cmd.Args,"111")
}

func ExistDir(dirname string) bool {
    fi, err := os.Stat(dirname)
    return (err == nil || os.IsExist(err)) && fi.IsDir()
}

func UnTar(dst, src string) error{
    command := "tar -zxvf " + src + " -C " + dst
    cmd := exec.Command("/bin/sh", "-c", command)
    cmdErr := cmd.Run()
    return cmdErr
}


func UnTar1(dst, src string) (err error) {
    // 打开准备解压的 tar 包
    fr, err := os.Open(src)
    if err != nil {
        return
    }
    print("111")
    defer fr.Close()

    // 将打开的文件先解压
    gr, err := gzip.NewReader(fr)
    if err != nil {
        return
    }
    defer gr.Close()
    print("2222222")
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
        // parse path cosid:
        dstFileDir := filepath.Join(dst, hdr.Name)
        print("333333333333333")
        // 根据 header 的 Typeflag 字段，判断文件的类型
        switch hdr.Typeflag {
        case tar.TypeDir: // 如果是目录时候，创建目录
            // 判断下目录是否存在，不存在就创建
            print("333334444444444\n")
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
            print("555555555555555\n")
            file, err := os.OpenFile(dstFileDir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
            if err != nil {
                return err
            }
            print("6666666666")
            n, err := io.Copy(file, tr)
            if err != nil {
                return err
            }
            print("777777777777")
            // 将解压结果输出显示
            util.Infof("成功解压： %s , 共处理了 %d 个字符\n", dstFileDir, n)

            // 不要忘记关闭打开的文件，因为它是在 for 循环中，不能使用 defer
            // 如果想使用 defer 就放在一个单独的函数中
            file.Close()
        }
    }

    return nil
}
/*
 *  path="cosid:1232134" 
 */
func GetCosData(cosid string) string{
  util.Info("*****************GetData from COS********************* ")
  // 根据cosid 拉取数据
  util.Info("getdata from toss..............")

  headers := map[string]string{}
  body := map[string]string{}
  params := map[string]string{}

  body["project_name"] = "pelivisio"
  body["app_id"] = cosid
  body["need_copy"] = "1"
  body["dest_path"] = "./testdata"

  url := env.TOSS_ADDR + env.TOSS_PREFIX
  //headers["STAFFNAME"] = "tq"
  util.Infof("url = %v",url)
  rep, _ := httprequest.Post(url, body, params, headers)

  // !!!make sure rep1 is map
  _, rep1 := httprequest.Parse2Json(rep)

  errcode := int(rep1["error"].(float64))
  if errcode == 0 {
    util.Infof("getdatasucess resp:%+v", rep1)
  }else{
    util.Infof("getdata false:%v", errcode)
    return ""
  }

  data := (rep1["data"]).(map[string] interface{})

  download_url := (data["download_url"]).(string)

  util.Infof("download_url = %s",download_url)
  resp, err := httprequest.Get(download_url,params,headers)
  if err != nil {
     panic(err)
  }
  defer resp.Body.Close()

  // 创建一个文件用于保存
  cos_name := cosid + ".tar.gz"
  out, err := os.Create(cos_name)
  if err != nil {
      panic(err)
  }
  defer out.Close()

    // 然后将响应流和文件流对接起来
  _, err = io.Copy(out, resp.Body)
  if err != nil {
     panic(err)
  }

  return cos_name
}

func SliceData2S3(sliceCnt int, cosid string, tarfilename string) string{ 


  // 解压当前目录 
  os.Mkdir(cosid,0755)
  UnTar(cosid,tarfilename)

  // 存放分割后的文件夹名字
  var listdestdir []string = make([]string, 0)
  // 创建rank个文件夹cosid
  subfile := []string{"Annotations","ImageSets","JPEGImages"}
  prefix_path := fmt.Sprintf("slice_%v_%v_",cosid,time.Now().Unix())
  for i:= 0; i < sliceCnt; i++ {
    filename := fmt.Sprintf("%v%v",prefix_path,i)
    os.Mkdir(filename,0755)
    for _,value := range subfile {
      os.MkdirAll(fmt.Sprintf("%v/%v",filename,value),0755)
      if value == "ImageSets"{
        os.MkdirAll(fmt.Sprintf("%v/%v/Main",filename,value),0755)
      }
    }

    listdestdir = append(listdestdir,filename)
  }

  // 原始文件
  srcfile := cosid

  CopySpliteData(srcfile,sliceCnt,listdestdir)

  //压缩
  for _,value := range listdestdir {
    slist := strings.Split(value,"/")
    fname := slist[len(slist)-1]
    zipname := fmt.Sprintf("%v.tar.gz",fname)
    Tar(value,zipname)
  }
  //上传

  return prefix_path
}

func CopySpliteData(sfile string, rank_size int, listdestdir []string ) bool {

  // get index cnt
  raw_filename := fmt.Sprintf("%v/VOCdevkit/VOC2007/ImageSets/Main/train.txt",sfile)
  cnt := GetlineCnt(raw_filename)
  if rank_size > cnt {
      return false
  }
  chunksize := cnt/rank_size
  util.Infof("cnt = %d, chunksize = %d",cnt,chunksize)

  // readdata2memory 
  slice := make([]string,0,cnt)
  file,err := os.Open(raw_filename)
	if err != nil{
		return false
	}
	defer file.Close()
	fd := bufio.NewReader(file)
	for {
		line,err := fd.ReadString('\n')
		if err!= nil{
			break
		}
    slice = append(slice,strings.TrimSpace(line))
	}

  // slice&write2file
  for i := 0; i < rank_size; i++ {
    istart := i*chunksize
    iend := (i+1)*chunksize
    s := slice[istart : iend]

    // write2newslicefile
    fname := fmt.Sprintf("%v/ImageSets/Main/train.txt",listdestdir[i])
    WriteMaptoFile(s,fname)

    otherfile := []string{ "back_train.txt","close_train.txt","trainval.txt","val.txt" }
    for _,value := range otherfile{
      // split文件
      srcfile := fmt.Sprintf("%v/VOCdevkit/VOC2007/ImageSets/Main/%v",sfile,value)
      destfile := fmt.Sprintf("%v/ImageSets/Main/%v",listdestdir[i],value)
      CopyFile(destfile,srcfile)
    }

    print("listdestdir = ",listdestdir[i])
    for _,value := range s{
      //copy img
      src_jpg := fmt.Sprintf("%v/VOCdevkit/VOC2007/JPEGImages/%v.jpg",sfile,value)
      dest_jpg := fmt.Sprintf("%v/JPEGImages/%v.jpg",listdestdir[i],value)
      CopyFile(dest_jpg,src_jpg)

      //copy xml
      src_xml := fmt.Sprintf("%v/VOCdevkit/VOC2007/Annotations/%v.xml",sfile,value)
      dest_xml := fmt.Sprintf("%v/Annotations/%v.xml",listdestdir[i],value)
      CopyFile(dest_xml,src_xml)
    }
  }

  return true
}
/*
 * split algorthm
 *  1 count
 *  2 chunksize = count/ranksize
 *  3 countid/chunksize -》1
 *                      -》2
 */
func Tar(src string , dst string) (err error) {
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
        hdr.Name = strings.TrimPrefix(hdr.Name, string(filepath.Separator)) 
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
        log.Printf("成功打包 %s ，共写入了 %d 字节的数据\n", fileName, n)

        return nil
    })
}

func UnTarold(dst, src string) (err error) {
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
        // parse path cosid:
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
            file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
            if err != nil {
                return err
            }
            n, err := io.Copy(file, tr)
            if err != nil {
                return err
            }
            // 将解压结果输出显示
            util.Infof("成功解压： %s , 共处理了 %d 个字符\n", dstFileDir, n)

            // 不要忘记关闭打开的文件，因为它是在 for 循环中，不能使用 defer
            // 如果想使用 defer 就放在一个单独的函数中
            file.Close()
        }
    }

    return nil
}
// 获取行数
func GetlineCnt(fileName string) int{

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

// 写文件
func WriteMaptoFile(m []string, filePath string) error {
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

func CopyFile(dstName, srcName string) (written int64, err error) {
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
func isZip(zipPath string) bool {
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
func unzip(archive, target string) error {
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
func Zipit(source, target, filter string) error {
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
        //log.Slogger.Errorf("*File close error: %s, file: %s", err.Error(), zipfile.Name())
		}
	}()

	//创建zip.Writer
	zw := zip.NewWriter(zipfile)

	defer func() {
		if err := zw.Close(); err != nil{
			  util.Infof("zipwriter close error: %s", err.Error())
        //log.Slogger.Errorf("zipwriter close error: %s", err.Error())
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
				//log.Slogger.Errorf("*File close error: %s, file: %s", err.Error(), file.Name())
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

type S3Manager struct {
	access_key string
	secret_key string
	end_point  string
  default_bucket string
  sess *session.Session
}

func (sm *S3Manager) Init() {

	sm.access_key = env.S3_ACCESSKEY
	sm.secret_key = env.S3_SECRETKEY
	sm.end_point = env.S3_ADDR //endpoint设置,不要动
  sm.default_bucket = env.S3_BUCKET

	sess,_ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(sm.access_key, sm.secret_key, ""),
		Endpoint:         aws.String(sm.end_point),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	})

  sm.sess = sess
}

//showallbuket
func (sm *S3Manager) ListBucket() {

	svc := s3.New(sm.sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		util.Infof("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	for _, b := range result.Buckets {
		fmt.Printf("%s\n", aws.StringValue(b.Name))
	}
}

// showall
func (sm *S3Manager) ListbucketFile(bucket string) {

	svc := s3.New(sm.sess)

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	result, err := svc.ListObjects(params)

	if err != nil {
		util.Infof("Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range result.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("List modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

// createbucket
func (sm *S3Manager) CreateBucketFile(bucket string) bool {

	svc := s3.New(sm.sess)

	params := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	_, err := svc.CreateBucket(params)

	if err != nil {
		util.Infof("Unable to create bucket %q, %v", bucket, err)
		return false
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		util.Infof("Error occurred while waiting for bucket to be created, %v", bucket)
		return false
	}

	fmt.Printf("Bucket %q successfully created\n", bucket)
	return true
}

// upload
func (sm *S3Manager) Upload(bucket string, filename string) {

	file, err := os.Open(filename)
	if err != nil {
		util.Infof("Unable to open file %v, %v",filename, err)
	}

	defer file.Close()

	uploader := s3manager.NewUploader(sm.sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		util.Infof("Unable to upload %q to %v, %v", filename, bucket, err)
	}

	util.Infof("Successfully uploaded %v to %v", filename, bucket)
}

// download
func (sm *S3Manager) Download(bucket string, filename string) {

	file, err := os.Create(filename)
	if err != nil {
		util.Infof("Unable to open file %q, %v", err)
	}

	defer file.Close()
	downloader := s3manager.NewDownloader(sm.sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
		})
	if err != nil {
		util.Infof("Unable to download item %v, %v", filename, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

}

// uploaddefault
func (sm *S3Manager) UploadDefaultbucket(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		util.Infof("Unable to open file %v, %v",filename, err)
	}

	defer file.Close()

	uploader := s3manager.NewUploader(sm.sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(sm.default_bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		util.Infof("Unable to upload %q to %v, %v", filename, sm.default_bucket, err)
	}

	util.Infof("Successfully uploaded %v to %v", filename, sm.default_bucket)
}
