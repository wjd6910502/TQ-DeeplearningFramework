package resourcemanager

import (
  "fmt"
  "time"
  "os"
  "io"
  "bufio"
  "strings"
  env "server/service/tq-service/env"
  httprequest "server/service/tq-service/httprequest"
  util "server/service/tq-service/util"
  )

/**
 * 基于voc数据集来分割 
 * 1加载数据从远程放到s3接口，
 * 1 从远程拉取数据
 * 2 对拉取到的数据进行切片
 * 3 将切片的数据放到s3/ceph上
 * 4 定义切片规则和命名规则，
 *    local 本地目录 local_rank
 *    s3拉取 slice_cosid_时间戳_rank
 *    参考：https://www.shuzhiduo.com/A/x9J2v0QWJ6/
 */
type LOADTYPE int32

const (
  LOADRESOURCE_LOCAL  LOADTYPE = 0
  LOADRESOURCE_COS    LOADTYPE = 1
  LOADRESOURCE_S3     LOADTYPE = 2
  LOADRESOURCE_HADOOP LOADTYPE = 3
  )

func PrepareData(sCnt int, train_path string,load_type int) string{
  if(load_type == int(LOADRESOURCE_COS) ){
    // parse path cosid:
    plist := strings.Split(train_path,":")
    cosid := plist[len(plist)-1]

    // Get data from cos
    tarfilename := GetCosData(cosid)

    // 解压与切割
    train_path := SliceData2S3(sCnt, cosid, tarfilename)
    return train_path
  } else if (load_type == int(LOADRESOURCE_S3)){

    util.Info("handle LOADRESOURCE_S3!!!!!!!!")
    return train_path
  } else if (load_type == int(LOADRESOURCE_HADOOP)){
    util.Info("handle LOADRESOURCE_HADOOP!!!!!!!!")
    return train_path
  } else {
    util.Info("cannot find loadtype!!!!!!!!")
    return train_path
  }
  return  train_path
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
  body["dest_path"] = "/data1/1"

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

func GetS3Data(path string){
  util.Info("*****************GetData from S3********************* ")
}

func GetHadoopData(path string){
  util.Info("*****************GetData from HADDOP********************* ")
}

func SliceData2S3(sliceCnt int, cosid string, tarfilename string) string{ 
  util.Infof("***************SliceData2S3**********************")
  fmgr := FileManager{}

  os.Mkdir(cosid,0755)
  err := fmgr.UnTar(cosid,tarfilename)
  if err != nil {
    util.Infof("Unzip File False.............")
  }
  util.Infof("************UnZip File Success*************")

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

  util.Infof("Copy split data........")
  srcfile := cosid

  // copy data 2 splitfile
  CopySpliteData(srcfile,sliceCnt,listdestdir)

  //压缩
  util.Info("Zip slice data.........")
  ziplist := make([]string,0)
  for _,value := range listdestdir {
    slist := strings.Split(value,"/")
    fname := slist[len(slist)-1]
    zipname := fmt.Sprintf("%v.tar.gz",fname)
    fmgr.Tar(value,zipname)
    ziplist = append(ziplist,zipname)
  }

  //上传
  util.Info("Upload s3 data..........")
  s3mgr := S3Manager{}
  s3mgr.Init()
  for _,value := range ziplist {
    s3mgr.UploadDefaultbucket(value)
  }
  return prefix_path
}

/*
 * split algorthm
 *  1 count
 *  2 chunksize = count/ranksize
 *  3 countid/chunksize -》1
 *                      -》2
 *                      -》3
 *                      -> ranksize
 *  param1: filename -> voc path  
 *  param2: ranksize   
 *  param3: cosid
 *  param4: listdir -> new construct small voc path
 *
*/
func CopySpliteData(sfile string, rank_size int, listdestdir []string ) bool {

  fm := FileManager{} 
  // get index cnt
  raw_filename := fmt.Sprintf("%v/VOCdevkit/VOC2007/ImageSets/Main/train.txt",sfile)
  cnt := fm.GetlineCnt(raw_filename)
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

    //  copy laebl_map
    labelfile := fmt.Sprintf("%v/VOCdevkit/label_map.json",sfile)
    dstfile := fmt.Sprintf("%v/label_map.json",listdestdir[i])
    fm.CopyFile(dstfile,labelfile)

    // write2newslicefile
    fname := fmt.Sprintf("%v/ImageSets/Main/train.txt",listdestdir[i])
    fm.WriteMaptoFile(s,fname)

    fname = fmt.Sprintf("%v/ImageSets/Main/val.txt",listdestdir[i])
    fm.WriteMaptoFile(s,fname)

    // label_map.json 2
    otherfile := []string{ "back_train.txt","close_train.txt","trainval.txt" }
    for _,value := range otherfile{
      // split文件
      srcfile := fmt.Sprintf("%v/VOCdevkit/VOC2007/ImageSets/Main/%v",sfile,value)
      destfile := fmt.Sprintf("%v/ImageSets/Main/%v",listdestdir[i],value)
      fm.CopyFile(destfile,srcfile)
    }

    print("listdestdir = ",listdestdir[i])
    for _,value := range s{
      //copy img
      src_jpg := fmt.Sprintf("%v/VOCdevkit/VOC2007/JPEGImages/%v.jpg",sfile,value)
      dest_jpg := fmt.Sprintf("%v/JPEGImages/%v.jpg",listdestdir[i],value)
      fm.CopyFile(dest_jpg,src_jpg)

      //copy xml
      src_xml := fmt.Sprintf("%v/VOCdevkit/VOC2007/Annotations/%v.xml",sfile,value)
      dest_xml := fmt.Sprintf("%v/Annotations/%v.xml",listdestdir[i],value)
      fm.CopyFile(dest_xml,src_xml)
    }
  }

  return true
}


