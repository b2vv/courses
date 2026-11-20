FROM node@sha256:0e2646ab57e77b7a384a0cb59b6815b182c77e1f9c3ba36bd8435bd9c4e6680e
RUN uname -m && node -p "process.version+' '+process.arch" && od -An -tx1 -j18 -N2 /usr/local/bin/node
