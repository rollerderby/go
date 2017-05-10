build() {
	echo "Building for $1-$2 to $3"
	GOOS=$1 GOARCH=$2 go install -v std || exit 1

	(go list -f '{{.Deps}}' $SERVER | tr "[" " " | tr "]" " " | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | grep --invert-match $BASE | GOOS=$1 GOARCH=$2 xargs -n 1 go install) || exit 1

	GOOS=$1 GOARCH=$2 go build -v -o $3 $SERVER || exit 1
	echo
}

ZIP=0
RELEASE=0
BASE=github.com/rollerderby/go
SERVER=$BASE/cmd/server

if [ "z$1" = "z-zip" ]; then ZIP=1; fi
if [ "z$1" = "z-release" ]; then RELEASE=1; ZIP=1; fi

VERSION=4.0.0
if [ $RELEASE -eq 0 ]; then VERSION=$VERSION-`date +%Y%m%d%H%M%S`; fi

echo Building Version $VERSION
echo

cat > cmd/server/version.go <<END
package main

const version = "$VERSION"
END

go get -u github.com/mjibson/esc || exit 1
rm -rf `go env GOPATH`/pkg
(go list -f '{{.Deps}}' $SERVER | tr "[" " " | tr "]" " " | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | grep --invert-match $BASE | xargs -n 1 go get -u) || exit 1
go install $BASE/cmd/buildStates || exit 1
go generate $BASE/cmd/server || exit 1
(go list -f '{{.Deps}}' $SERVER | tr "[" " " | tr "]" " " | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | grep $BASE | xargs -n 1 go generate) || exit 1

if [ $ZIP -eq 0 ]; then
	mkdir -p bin
	PLATFORM=$(uname -s)
	EXT=""
	case $PLATFORM in
		CYGWIN*|MINGW32*|MSYS*)
			EXT=".exe"
			;;
	esac
	rm -f ./bin/scoreboard$EXT
	build `go env GOOS` `go env GOARCH` "bin/scoreboard$EXT"
else
	rm -f scoreboard-*
	mkdir -p release
	rm -f release/scoreboard-$VERSION.zip

	mkdir -p crg-scoreboard_$VERSION
	cp -r html crg-scoreboard_$VERSION
	cp AUTHORS LICENSE crg-scoreboard_$VERSION

	build "linux" "386" "crg-scoreboard_$VERSION/scoreboard-Linux-32"
	build "linux" "amd64" "crg-scoreboard_$VERSION/scoreboard-Linux-64"
	build "windows" "386" "crg-scoreboard_$VERSION/scoreboard-Windows-32.exe"
	build "windows" "amd64" "crg-scoreboard_$VERSION/scoreboard-Windows-64.exe"
	build "darwin" "386" "crg-scoreboard_$VERSION/scoreboard-MacOS-32"
	build "darwin" "amd64" "crg-scoreboard_$VERSION/scoreboard-MacOS-64"

	echo Zipping to release/crg-scoreboard_$VERSION.zip
	zip -qr release/crg-scoreboard_$VERSION.zip crg-scoreboard_$VERSION
	rm -rf crg-scoreboard_$VERSION
fi
