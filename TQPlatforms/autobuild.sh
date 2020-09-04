if [ $# -lt 2 ]; then
	echo "########### help #########"
	echo "need mode state"
	echo "./autobuild.sh start-local"
	echo "./autobuild.sh start-remote"
	echo "########### help ##########"
fi

# local test
st="start-local"
if [ "$1" == "$st" ]; then 
	docker build . -f ./Dockerfile -t dmlp-service
	docker run -it -p 8812:5050 dmlp-service /bin/bash
fi

sr="start-remote"
if [ "$1" = "$sr" ]; then
	#local package
	docker tag
	docker build . -f ./Dockerfile -t dmlp-service
	id=`docker images | grep dmlp-service | awk -F' ' 'NR==1{print $3}' | awk -F' ' '{print $1}'`
	if [ $? -ne 0 ]; then
	    echo "failed"
		exit
	fi
	echo $id
	echo "build succeed!!!!!!!!!!!!!!!!!"
	
	#docker images
	res=`docker images | grep "dev.oa.com/turinglab/dmlp-service" | sort -r | awk -F' ' 'NR==1{print $2}'`
	if [ $? -ne 0 ]; then
	    echo "failed"
		exit
	fi
	
	ver=1
	if [ "$res" != "" ]; then
		a=0.1
		ver=$(echo "scale=2;$res+$a"|bc)
	fi
  
  echo "1111"
	echo $ver
	
	docker tag $id dev.oa.com/turinglab/dmlp-service:$ver
	docker login dev.oa.com
	docker push dev.oa.com/turinglab/dmlp-service:$ver
	
	# delete noused iamge
	docker rmi -f  `docker images | grep '<none>' | awk '{print $3}'`
	docker images | grep dlmp-service
fi
