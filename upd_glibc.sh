#!/bin/bash

ubuntuneedversion="eoan"

glibver=$((dpkg -s libc6 || rpm -qi glibc) 2>/dev/null | grep -i version | sed -r 's/.*([0-9]+\.[0-9]+).*|.*/\1/' | uniq | grep -v '^$')
ver=0
subver=0

if [ -z $glibver ]
then
	echo "glibc not found!"
else
	echo "glibc version: $glibver"

	ver=$(echo "$glibver" | cut -d. -f1)
	subver=$(echo "$glibver" | cut -d. -f2)
fi

if (( ver * 1000 + subver < 2028 ))
then
	echo "need to upgrade glibc!"

	os=$(cat /etc/os-release | grep -iE "^id=" | cut -d= -f2 | sed 's/"//g')
	osver=$(cat /etc/os-release | grep -i "version_codename=" | cut -d= -f2 | sed 's/"//g')
	echo "os: $os $osver"

	case "$os" in
		ubuntu )
			echo "upgrading ubuntu repo..."
			if [ -z "$(cat /etc/apt/sources.list | grep -iE "^deb http://[a-z.]+/ubuntu $ubuntuneedversion main ")" ]
			then
				echo "adding $ubuntuneedversion repo"
				sudo cat <<EOT | sudo tee /etc/apt/preferences.d/glibc.pref > /dev/null
Package: *
Pin: release n=$ubuntuneedversion
Pin-Priority: -10

Package: libc6
Pin: release n=$ubuntuneedversion
Pin-Priority: 500

Package: libc-*
Pin: release n=$ubuntuneedversion
Pin-Priority: 500

Package: ubuntu-minimal
Pin: release n=$ubuntuneedversion
Pin-Priority: 500
EOT
				if [ $? -ne 0 ]
				then
					echo "cannot write rule for packages, exiting"
					exit 1
				fi
				sudo cat <<EOT | sudo tee -a /etc/apt/sources.list > /dev/null
deb http://archive.ubuntu.com/ubuntu ${ubuntuneedversion} main universe restricted multiverse
deb-src http://archive.ubuntu.com/ubuntu ${ubuntuneedversion} main universe restricted multiverse
deb http://archive.ubuntu.com/ubuntu ${ubuntuneedversion}-security main universe restricted multiverse
deb-src http://archive.ubuntu.com/ubuntu ${ubuntuneedversion}-security main universe restricted multiverse
deb http://archive.ubuntu.com/ubuntu ${ubuntuneedversion}-updates multiverse
EOT
			fi
			sudo apt update > upd.log
			# echo "-----------------------------------------" >> upd.log
			sudo apt install libc6 #>> upd.log
			sudo apt install -f #>> upd.log

			newglibver=$((dpkg -s libc6 || rpm -qi glibc) 2>/dev/null | grep -i version | sed -r 's/.*([0-9]+\.[0-9]+).*|.*/\1/' | uniq | grep -v '^$')
			if [ "$newglibver" == "$glibver" ]
			then
				echo "not updated :("
				echo ""
				apt-cache policy libc6
			else
				echo "new glibc version: $newglibver"
			fi
			;;
		centos )
			echo "not implemented yet :("
			;;
		* )
			echo "unknown os to do :("
			;;
	esac
fi