package main // github.com/dnywonnt/launcher-prestarter-unix

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dnywonnt/launcher-prestarter-unix/utils"
	log "github.com/sirupsen/logrus"
)

const (
	PROJECT_LOGO = `	
██╗░░░██╗░█████╗░██╗░░░██╗██████╗░  ██████╗░██████╗░░█████╗░░░░░░██╗███████╗░█████╗░████████╗
╚██╗░██╔╝██╔══██╗██║░░░██║██╔══██╗  ██╔══██╗██╔══██╗██╔══██╗░░░░░██║██╔════╝██╔══██╗╚══██╔══╝
░╚████╔╝░██║░░██║██║░░░██║██████╔╝  ██████╔╝██████╔╝██║░░██║░░░░░██║█████╗░░██║░░╚═╝░░░██║░░░
░░╚██╔╝░░██║░░██║██║░░░██║██╔══██╗  ██╔═══╝░██╔══██╗██║░░██║██╗░░██║██╔══╝░░██║░░██╗░░░██║░░░
░░░██║░░░╚█████╔╝╚██████╔╝██║░░██║  ██║░░░░░██║░░██║╚█████╔╝╚█████╔╝███████╗╚█████╔╝░░░██║░░░
░░░╚═╝░░░░╚════╝░░╚═════╝░╚═╝░░╚═╝  ╚═╝░░░░░╚═╝░░╚═╝░╚════╝░░╚════╝░╚══════╝░╚════╝░░░░╚═╝░░░
	`

	PROJECT_NAME = "Your Project"
	LAUNCHER_URL = "https://urltoyourproject.com/Launcher.jar"
	PROJECT_HELLO_TEXT = "Your project hello text"

	JRE_LINUX_X64_URL = "https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse?project=jdk"
	JFX_LINUX_X64_URL = "https://download2.gluonhq.com/openjfx/21.0.1/openjfx-21.0.1_linux-x64_bin-sdk.zip"

	JRE_MACOS_X64_URL = "https://api.adoptium.net/v3/binary/latest/21/ga/mac/x64/jre/hotspot/normal/eclipse?project=jdk"
	JFX_MACOS_X64_URL = "https://download2.gluonhq.com/openjfx/21.0.1/openjfx-21.0.1_osx-x64_bin-sdk.zip"
)

// хук форматтера логов
type formatterHook struct {
	Writer    io.Writer
	LogLevels []log.Level
	Formatter log.Formatter
}

// метод записи в лог
func (hook *formatterHook) Fire(entry *log.Entry) error {
	line, lerr := hook.Formatter.Format(entry)
	if lerr != nil {
		return lerr
	}
	_, lerr = hook.Writer.Write(line)
	return lerr
}

// метод хука для получения уровней логов
func (hook *formatterHook) Levels() []log.Level {
	return hook.LogLevels
}

func main() {
	/* по умолчанию, отправлять логи в никуда
	мы сами позже установим необходимые хуки */
	log.SetOutput(io.Discard)
	log.SetLevel(log.DebugLevel)

	log.AddHook(&formatterHook{ // хук в Stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.FatalLevel,
		},
		Formatter: &log.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
		},
	})

	logBuffer := &bytes.Buffer{}
	log.AddHook(&formatterHook{ // хук в буффер
		Writer: logBuffer,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
		Formatter: &log.JSONFormatter{},
	})
	defer logBuffer.Reset() // сбрасываем буффер при завершении работы

	log.Debugf("ОС: %v (%v)", runtime.GOOS, runtime.GOARCH) // логгируем информацию об ОС юзера

	fmt.Println(PROJECT_LOGO) // выводим лого проекта

	var dirs [3]string
	var filePathes [3]string
	var urls [3]string

	uhdir, uhderr := os.UserHomeDir() // получаем домашнюю директорию пользователя
	if uhderr != nil {
		log.Fatal(uhderr)
	}

	baseDir := filepath.Join(uhdir, ".minecraftlauncher", PROJECT_NAME) // базовая директория
	/*  проверяем, существует ли базовая директория и выводим сообщение
	приветствия, если нет, и создаем ее*/
	if _, osserr := os.Stat(baseDir); os.IsNotExist(osserr) {
		fmt.Print(PROJECT_HELLO_TEXT)
		fmt.Scanln()

		if mkderr := os.Mkdir(baseDir, 0755); mkderr != nil {
			log.Fatal(mkderr)
		}
		log.Infof("Директория '%v' создана.", baseDir)
	} else {
		log.Infof("Директория '%v' найдена.", baseDir)
	}
	dirs[0] = filepath.Join(baseDir, "java")                            // директория JAVA
	dirs[1] = filepath.Join(dirs[0], "jdk-21.0.1+12-jre")               // директория JRE
	dirs[2] = filepath.Join(dirs[0], "javafx-sdk-21.0.1")               // директория JFX
	urls[0] = LAUNCHER_URL                                              // ссылка для скачивания лаунчера
	filePathes[0] = filepath.Join(baseDir, filepath.Base(LAUNCHER_URL)) // путь до лаунчера

	// проверяем ОС
	switch runtime.GOOS {
	case "linux":
		// ссылки для скачивания JRE и JFX linux x64
		urls[1] = JRE_LINUX_X64_URL
		urls[2] = JFX_LINUX_X64_URL

		// пути до архивов JRE и JFX linux x64
		filePathes[1] = filepath.Join(baseDir, "jre_linux_x64.tar.gz")
		filePathes[2] = filepath.Join(baseDir, "jfx_linux_x64.zip")
	case "darwin":
		// ссылки для скачивания JRE и JFX macos x64
		urls[1] = JRE_MACOS_X64_URL
		urls[2] = JFX_MACOS_X64_URL

		// пути до архивов JRE и JFX macos x64
		filePathes[1] = filepath.Join(baseDir, "jre_macos_x64.tar.gz")
		filePathes[2] = filepath.Join(baseDir, "jfx_macos_x64.zip")
	}

	// перебираем все пути до файлов и скачиваем, если их не существует
	for i, filePath := range filePathes {
		if _, osserr := os.Stat(filePath); os.IsNotExist(osserr) {
			dferr := utils.DownloadFile(urls[i], filePath, fmt.Sprintf("Загрузка '%v':", filepath.Base(filePath)), "launcher-prestarter-unix")
			if dferr != nil {
				log.Fatal(dferr)
			}
			log.Infof("Файл '%v' загружен.", filepath.Base(filePath))
		} else {
			log.Infof("Файл '%v' найден.", filepath.Base(filePath))
		}
	}

	// распаковываем архивы, если не распакованы
	for i, dir := range dirs {
		// проверяем существование директорий JAVA
		if _, osserr := os.Stat(dir); os.IsNotExist(osserr) {
			// проверяем тип архива
			switch ext := filepath.Ext(filePathes[i]); ext {
			case ".gz", ".zip":
				var uerr error
				if ext == ".gz" {
					uerr = utils.UnpackTarGz(filePathes[i], dirs[0], fmt.Sprintf("Распаковка '%v':", filepath.Base(filePathes[i])))
				} else if ext == ".zip" {
					uerr = utils.UnzipFile(filePathes[i], dirs[0], fmt.Sprintf("Распаковка '%v':", filepath.Base(filePathes[i])))
				}
				if uerr != nil {
					log.Fatal(uerr)
				}
				log.Infof("Архив '%v' распакован.", filepath.Base(filePathes[i]))
			}
		} else {
			log.Infof("Директория '%v' найдена.", dir)
		}
	}

	// проверяем существование модулей JavaFx в директории JRE
	if jfxIsExists, federr := utils.FilesExistInDirectory(filepath.Join(dirs[1], "lib"), "javafx"); !jfxIsExists {
		// копируем файлы из директории JFX в директорию JRE с заменой
		if mferr := utils.CopyFiles(dirs[2], dirs[1], fmt.Sprintf("Копирование файлов '%v':", filepath.Base(dirs[2]))); mferr != nil {
			log.Fatal(mferr)
		}
		log.Infof("Модули JavaFX скопированы в '%v'.", dirs[1])
	} else if federr != nil {
		log.Fatal(federr)
	} else {
		log.Info("Модули JavaFX найдены.")
	}

	// запускаем лаунчер
	cmd := exec.Command(filepath.Join(dirs[1], "bin", "java"), "-jar", filePathes[0])
	if cmderr := cmd.Run(); cmderr != nil {
		log.Fatal(cmderr)
	}
	log.Info("Лаунчер запущен.")

	// создаем или открываем лог файл
	logFile, lferr := os.OpenFile(filepath.Join(baseDir, "launcher-prestarter-unix.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if lferr != nil {
		log.Fatal(lferr)
	}
	// записываем логи и закрываем файл при завершении работы
	defer logFile.Close()
	defer logFile.Write(logBuffer.Bytes())
}
