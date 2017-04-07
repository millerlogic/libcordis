
go build -buildmode=c-shared \
    -ldflags '-s -w' \
    -o libcordis.so github.com/millerlogic/libcordis
