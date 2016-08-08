#!/usr/bin/env bash

killall gur
sleep 3
rm -f /tmp/gur.log
SYSTEM=`uname -a | cut -d' ' -f1`
if [ $SYSTEM = "Darwin" ]; then
  # echo "os x"
  mkdir -p ~/Library/UR/
  ln -s ~/Library/UR/ ~/.ur
else
  # echo "linux"
  mkdir -p ~/.ur
fi
cd ~/.ur
rm -rf `ls | grep -v 'nodekey\|keystore'` # these files define node urls and account key pairs
cd ~
BRANCH_OR_COMMIT=master
if [ -d ~/go-ur ]; then
  echo "clone of go-ur already exists at ~/go-ur"
  cd ~/go-ur
  git fetch
  git checkout $BRANCH_OR_COMMIT
  git reset --hard origin/$BRANCH_OR_COMMIT
  echo "did a hard reset to origin/$BRANCH_OR_COMMIT"
else
  echo "cloning go-ur to ~/go-ur and checking out commit $BRANCH"
  git clone https://github.com/urcapital/go-ur.git
  cd ~/go-ur
  git checkout $BRANCH_OR_COMMIT
fi
make gur > /tmp/makegur.log
echo "built gur"
