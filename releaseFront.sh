#!/bin/bash
srcdir=~/golang/src/opsv2.cn/opsv2/frontend
dest=/data/www/opsv2
rm -rf /data/www/opsv2/*
cd $srcdir
function die() {
  echo $1
  exit 1
}
if [ "$1" = "release" ]; then
  [ -d $dest/statics ] || mkdir -p $dest/statics
  [ -d $dest/html ] || mkdir $dest/html
  [ -d $dest/statics/fonts ] || mkdir -p $dest/statics/fonts
  [ -d $dest/statics/img ] || mkdir -p $dest/statics/img
  cp -r statics/fonts/* $dest/statics/fonts/
  cp -r statics/img/* $dest/statics/img/
  #awk: if (i > 0 && substr($0,i+2,13) != "statics/myjs/") print substr($0,i+2,RLENGTH-2);
  for item in `awk 'BEGIN{common=0}{
    if ($0 ~ /<!--/) {common=1};
    if ($0 ~ /-->/) {common=0;next};
    if (common == 1) next;
    i = match($0,"=\"statics/[^\"]+")
    if (i > 0 ) print substr($0,i+2,RLENGTH-2);
    }END{
      print "statics/layer/theme/default/layer.css";
      print "statics/fonts/fonts.css";
    }' index.html`; do
      path=`echo $item | sed -e 's/\/[^\/]*$//'`
      [ -d $dest/$path ] || mkdir -p $dest/$path
      cp $item $dest/$item || echo "copy $item failed"
      echo $item | grep -E '\.(js|css)$' > /dev/null 
      if [ $? -eq 0 ]; then
        sed -i -e 's/sourceMappingURL.*//' $dest/$item
        gzip -f -k $dest/$item 
      fi
  done
  sed -i -E -e 's#https://fonts.googleapis.com/[^)]+#../fonts/fonts.css#' $dest/statics/css/AdminLTE.min.css
  gzip -f -k $dest/statics/css/AdminLTE.min.css
  cp -r statics/plugins/font-awesome/* $dest/statics/plugins/font-awesome
  [ -d $dest/statics/bootstrap/fonts ] || mkdir -p $dest/statics/bootstrap/fonts
  cp -r statics/bootstrap/fonts/* $dest/statics/bootstrap/fonts/
  cp -r statics/plugins/iCheck/* $dest/statics/plugins/iCheck/
  for html in `awk '{ if ($0 ~/[^/]templateUrl/) {i=match($2,"[^/]+.html");tt=substr($2,i,RLENGTH);dd[tt]=1}}END{for (ff in dd) print ff}' statics/myjs/app.js`; do
    cp html/$html $dest/html/$html
    gzip -f -k $dest/html/$html
  done
  cp index.html $dest/
  #awk -v minjs='<script src="statics/myjs/opsv2.min.js"></script>' 'BEGIN{
  #  ops=0
  #}{
  #  if ($1 ~ /<!--opsv2-->/) {
  #    if (ops == 0){
  #      ops=1;
  #      print minjs;
  #      next;
  #    } else {
  #      ops=0;
  #      next;
  #    }
  #  };
  #  if (ops==0) print $0
  #}' index.html > $dest/index.html
  gzip -f -k $dest/index.html
  #cp statics/myjs/app.js ~/test/
  sed -i -e 's#http://console.zdops.com#https://ops.daledi.cn#' $dest/statics/myjs/app.js
  #java -jar ~/closure-compiler-v20180402.jar \
  #  --angular_pass \
  #  --js=/home/dale/test/app.js \
  #  --js=statics/myjs/services/service.js \
  #  --js=statics/myjs/template/template.js \
  #  --js=statics/myjs/injectors/injector.js \
  #  --js=statics/myjs/controllers/projectController.js \
  #  --js=statics/myjs/controllers/cmdbController.js \
  #  --js=statics/myjs/controllers/devicesController.js \
  #  --js=statics/myjs/controllers/jobsController.js \
  #  --js=statics/myjs/controllers/userController.js \
  #  --js=statics/myjs/controllers/RoleController.js \
  #  --js=statics/myjs/controllers/ApptplController.js \
  #  --js=statics/myjs/controllers/baseController.js \
  #  --compilation_level SIMPLE_OPTIMIZATIONS \
  #  --create_source_map $dest/statics/myjs/opsv2.map \
  #  --js_output_file $dest/statics/myjs/opsv2.min.js || die "js cc failed"
  #gzip -f -k $dest/statics/myjs/opsv2.min.js
fi
