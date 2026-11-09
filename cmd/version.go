package cmd

// Version 是程序版本号，发布流程通过
// -ldflags "-X image-poller/cmd.Version=<version>" 注入。
var Version = "dev"
