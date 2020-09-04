if [ $# -lt 2 ]; then
	echo "########### help #########"
	echo "****** need mode state******"
	echo "./autobuild.sh start-local"
	echo "./autobuild.sh start-remote"
	echo "########### help ##########"
fi

# local test
st="start-local"
if [ "$1" == "$st" ]; then 
	docker build . -f ./Dockerfile -t hvd-platform
	docker run -it -p 8812:5050 hvd-platform /bin/bash
fi

sr="start-remote"
if [ "$1" = "$sr" ]; then
	#local package
	docker tag
	docker build . -f ./Dockerfile -t dmlp-framework
	id=`docker images | grep dmlp-framework | awk -F' ' 'NR==1{print $3}' | awk -F' ' '{print $1}'`
	if [ $? -ne 0 ]; then
	    echo "failed"
		exit
	fi
	echo $id
	echo "build succeed!!!!!!!!!!!!!!!!!"
	
	#docker images
	res=`docker images | grep "hub.oa.com/turinglab/dmlp-framework" | sort -r | awk -F' ' 'NR==1{print $2}'`
	if [ $? -ne 0 ]; then
	    echo "failed"
		exit
	fi
	
	ver=1
	if [ "$res" != "" ]; then
		a=0.1
		ver=$(echo "scale=2;$res+$a"|bc)
	fi
								
	echo $ver
	
	docker tag $id hub.oa.com/turinglab/dmlp-framework:m-v0.3.0.0
	docker login hub.oa.com
	docker push hub.oa.com/turinglab/dmlp-framework:m-v0.3.0.0
	
	# delete noused iamge
	docker rmi -f  `docker images | grep '<none>' | awk '{print $3}'`
	docker images | grep dmlp-framework
fi
