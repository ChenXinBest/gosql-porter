package tools

import (
	"archive/zip"
	"fmt"
	"gosql/internal/config"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	// 检查 mysqlshell 命令是否可用
	cmd := exec.Command("mysqlsh", "--version")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("请先安装 mysqlshell，并添加到环境变量（mysqlsh --version 不可用）: ", err)
	}
	fmt.Println("检测到 mysqlshell：", string(out))
}

// 检查连接是否可用
func CheckConnection(cnf config.Config) bool {
	mode := cnf.Options.Mode
	sourceDb := buildUri(cnf.DBSettings.Source)
	targetDb := buildUri(cnf.DBSettings.Target)
	switch mode {
	case 1:
		// 仅导入
		return shellTestLink(sourceDb)
	case 2:
		// 仅导出
		return shellTestLink(targetDb)
	default:
		// 导入导出
		return shellTestLink(sourceDb) && shellTestLink(targetDb)
	}
}

// 测试连接
func shellTestLink(dbUriParam string) bool {
	cmd := exec.Command(
		"mysqlsh",
		dbUriParam,
		"--js",
		"-e",
		"shell.status()",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("连接【%s】出错：%s\n%s", dbUriParam, err, string(out))
		return false
	}
	res := string(out)
	fmt.Printf("连接信息输出如下：\n%s\n", res)
	if strings.Contains(res, "version") {
		fmt.Printf("连接成功！\n")
		return true
	} else {
		fmt.Println("未检测到版本信息！")
		return false
	}
}

// 将文件导入到数据库
func LoadDump(cnf config.Config) {
	var jsonFlag bool = false
	uri := buildUri(cnf.DBSettings.Target)
	dumpTo := cnf.Options.DumpTo
	// 判断导出文件夹是否存在且有必要的json和sql文件
	info, err := os.Stat(dumpTo)
	if os.IsNotExist(err) {
		log.Fatalln("文件夹不存在！")
	} else if err != nil {
		log.Fatalln("读取文件夹出现错误！", err)
	}
	if !info.IsDir() {
		log.Fatalln("导出目录不是文件夹！", dumpTo)
	}
	// 扫描文件夹
	filepath.WalkDir(dumpTo, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("访问路径出错：%q,%v\n", path, err)
		}

		relPath, _ := filepath.Rel(dumpTo, path)

		if filepath.Ext(relPath) == ".json" {
			jsonFlag = true
			return nil
		}
		return nil
	})
	if !jsonFlag {
		log.Fatalln("未扫描到json文件，请确认导出目录状态！")
	}
	// 删库
	DelSchemas(cnf)
	// 导入数据库
	command := fmt.Sprintf("util.loadDump('%s',{threads: %d})", cnf.Options.DumpTo, cnf.Options.Threads)
	cmd := exec.Command(
		"mysqlsh",
		uri,
		"--js",
		"-e",
		command,
	)
	fmt.Println("正在执行命令：", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("导入数据库出错: %s\n错误详情: %v", string(out), err)
	}
	fmt.Println("导入成功！\n", string(out))
}

// 导出数据库到目录
func DumpSql(cnf config.Config) {
	dbSetting := cnf.DBSettings.Source
	dumpPath := cnf.Options.DumpTo
	// 判断导出目录是否存在
	_, err := os.Stat(dumpPath)
	if os.IsNotExist(err) {
		// 目录不存在，创建目录
		fmt.Println("目录不存在，尝试递归创建目录：", dumpPath)
		if err := os.MkdirAll(dumpPath, 0755); err != nil {
			fmt.Println("创建目录失败：", err)
			return
		}
		fmt.Println("目录创建成功：", dumpPath)
	} else {
		// 删除文件夹下所有东西
		// 读取目录内容
		entries, err := os.ReadDir(dumpPath)
		if err != nil {
			log.Fatalln("读取目录失败：", err)
			return
		}

		// 逐个删除
		for _, entry := range entries {
			path := filepath.Join(dumpPath, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Println("删除失败：", path, err)
				return
			}
		}
		fmt.Println("目录内容已清空：", dumpPath)
	}
	uri := buildUri(dbSetting)
	schemas := strings.Join(cnf.Options.Databases, "','")
	schemas = "'" + schemas + "'"
	command := fmt.Sprintf("util.dumpSchemas([%s], '%s', {threads: %d})", schemas, cnf.Options.DumpTo, cnf.Options.Threads)
	fmt.Println("开始导出数据库……")
	// 构造指令
	cmd := exec.Command(
		"mysqlsh",
		uri,
		"--js",
		"-e",
		command,
	)

	// 打印命令用于调试
	// fmt.Println("执行命令:", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln("导出数据库出错！\n", string(out), err)
	}
	fmt.Println("导出成功！", string(out))
}

func DelSchemas(cnf config.Config) {
	if len(cnf.Options.Databases) == 0 {
		return
	}
	uri := buildUri(cnf.DBSettings.Target)
	for _, db := range cnf.Options.Databases {
		fmt.Println("正在删除数据库：", db)
		cmd := exec.Command(
			"mysqlsh",
			uri,
			"--sql",
			"-e",
			fmt.Sprintf("drop database if exists `%s`;", db),
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalln("删库出错！", string(out), err)
		}
	}
}

// 备份导出文件
func SaveDump(cnf config.Config) {
	timeStr := fmt.Sprint(time.Now().Format("20060102150405"))
	fileName := "backup_" + timeStr + ".zip"
	filePath := filepath.Join(cnf.Options.SaveTo, fileName)
	dumpPath := cnf.Options.DumpTo
	info, err := os.Stat(dumpPath)
	if os.IsNotExist(err) {
		log.Fatalln("请先导出到文件夹：", dumpPath)
	} else if err != nil {
		log.Fatalln("获取文件夹状态失败：", err)
	}
	if !info.IsDir() {
		log.Fatalln("导出目录不是一个目录：", dumpPath)
	}

	// 确保保存目录存在
	saveDir := filepath.Dir(filePath)
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			log.Fatalln("创建保存目录失败：", err)
		}
	}

	// 创建zip文件
	zipFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalln("创建zip文件失败：", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历dumpPath目录，将所有文件添加到zip中
	err = filepath.Walk(dumpPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(dumpPath, path)
		if err != nil {
			return err
		}

		// 创建zip文件头
		zipHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		zipHeader.Name = relPath

		// 如果是目录，确保路径以分隔符结尾
		if info.IsDir() {
			zipHeader.Name += "/"
		} else {
			zipHeader.Method = zip.Deflate
		}

		// 写入文件头
		writer, err := zipWriter.CreateHeader(zipHeader)
		if err != nil {
			return err
		}

		// 如果是文件，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalln("压缩文件失败：", err)
	}

	fmt.Println("备份成功！文件保存至：", filePath)
}

func buildUri(dbSetting config.Database) string {
	// 对用户信息进行url编码
	pwd := url.QueryEscape(dbSetting.Passwd)
	username := url.QueryEscape(dbSetting.Username)
	return fmt.Sprintf("--uri=%s:%s@%s:%d",
		username,
		pwd,
		dbSetting.Host,
		dbSetting.Port,
	)
}
