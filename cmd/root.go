/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gosql/internal/config"
	"gosql/internal/tools"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var configPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gosql",
	Short: "对mysqlshell指令的包装脚本",
	Long:  `对mysqlshell的二次包装，以实现便捷的数据库导出导入功能`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	if configPath == "" {
		configPath = "config.yml"
	}
	cnf, err := config.Load(configPath)
	if err != nil {
		log.Fatalln("读取配置文件出错：", err)
	}
	if cnf.Options.DumpTo == "" {
		cnf.Options.DumpTo = "dumpSql"
	}
	if len(cnf.Options.Databases) == 0 {
		log.Fatalln("配置的[options-databases]数据库列表不得为空！")
	}

	// 检查mysqlsh连接
	if !tools.CheckConnection(*cnf) {
		log.Fatalln(fmt.Sprintln("mysqlsh无法完成连接！"))
	}
	mode := cnf.Options.Mode
	if mode == 0 || mode == 1 {
		// 需要导出
		tools.DumpSql(*cnf)
		// 需要备份
		if cnf.Options.SaveTo != "" {
			tools.SaveDump(*cnf)
		}
	}
	if mode == 0 || mode == 2 {
		// 需要导入
		tools.LoadDump(*cnf)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gosql.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "获取帮助信息")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "指定配置文件路径，默认当前路径下的config.yml")
}
