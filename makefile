# Basic go commands
.PHONY: all build clean run default help

GOCMD=go
GOBUILD=$(GOCMD) build 
GOCLEAN=$(GOCMD) clean

MAINFILE=main.go
# Binary names
BINARY_NAME=webhook
BINARY_UNIX=webhook
BINARY_ARM=webhook


default:
		$(GOBUILD) -o $(BINARY_NAME) 
main:
		$(GOBUILD) -o $(BINARY_NAME) -o $(BINARY_UNIX) $(MAINFILE)
amd:
		GOOS=linux GOARCH=amd64 $(GOBUILD)  -o $(BINARY_UNIX) $(MAINFILE)

#win:
#		GOOS=windows GOARCH=amd64 $(GOBUILD)  -o $(BINARY_WIN) $(MAINFILE)

arm:
		GOOS=linux GOARCH=arm64 $(GOBUILD)  -o $(BINARY_ARM) $(MAINFILE)

clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		$(GOBUILD) -o $(BINARY_NAME) -v 
		./$(BINARY_NAME)
help:
		@echo "make 编译生成二进制文件"
		@echo "make build 编译生成linux amd64二进制文件"
		@echo "make clean 清理中间目标文件"
		@echo "make run 编译并直接运行程序"

