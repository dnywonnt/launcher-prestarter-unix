package utils // github.com/dnywonnt/launcher-prestarter-unix/utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/schollz/progressbar/v3"
)

// функция скачивания файла с прогрессбаром
func DownloadFile(url, outPath, barText, userAgent string) error {
	req, reqerr := http.NewRequest("GET", url, nil)
	if reqerr != nil {
		return reqerr
	}
	// устанавливаем юзер-агент
	req.Header.Set("User-Agent", userAgent)
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return resperr
	}
	defer resp.Body.Close()

	// создаем файл для записи
	outFile, oferr := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0744)
	if oferr != nil {
		return oferr
	}
	defer outFile.Close()

	// инициализируем прогрессбар
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		barText,
	)
	if _, iocerr := io.Copy(io.MultiWriter(outFile, bar), resp.Body); iocerr != nil {
		return iocerr
	}

	return nil
}

// функция распаковки tar.gz архива с прогрессбаром
func UnpackTarGz(tarGzPath, destPath, barText string) error {
	// открываем tar.gz архив для чтения
	tarGzFile, tgferr := os.Open(tarGzPath)
	if tgferr != nil {
		return tgferr
	}
	defer tarGzFile.Close()

	// создаем gzip.Reader для чтения сжатых данных
	gzipReader, grerr := gzip.NewReader(tarGzFile)
	if grerr != nil {
		return grerr
	}
	defer gzipReader.Close()

	// создаем tar.Reader для чтения tar архива
	tarReader := tar.NewReader(gzipReader)

	// итерируемся по файлам в архиве
	for {
		header, herr := tarReader.Next()

		switch {
		case herr == io.EOF:
			// конец архива
			return nil
		case herr != nil:
			// произошла ошибка
			return herr
		case header == nil:
			// пропускаем записи без заголовка
			continue
		}

		// формируем полный путь к файлу внутри архива
		targetFilePath := filepath.Join(destPath, header.Name)

		// проверяем тип записи
		switch header.Typeflag {
		case tar.TypeDir:
			// cоздаем директории, если они отсутствуют
			if mkderr := os.MkdirAll(targetFilePath, 0755); mkderr != nil {
				return mkderr
			}
		case tar.TypeReg, tar.TypeSymlink:
			// создаем файл и копируем данные из архива
			file, ferr := os.Create(targetFilePath)
			if ferr != nil {
				return ferr
			}
			// устанавливаем права
			if fcerr := file.Chmod(0744); fcerr != nil {
				return fcerr
			}
			defer file.Close()

			// инициализируем прогрессбар
			bar := progressbar.DefaultBytes(
				header.Size,
				barText+" "+header.Name,
			)
			if _, iocerr := io.Copy(io.MultiWriter(file, bar), tarReader); iocerr != nil {
				return iocerr
			}
		}
	}
}

// функция распаковки zip архива с прогрессбаром
func UnzipFile(zipFile, destDir, barText string) error {
	// открываем zip-архив для чтения
	reader, rerr := zip.OpenReader(zipFile)
	if rerr != nil {
		return rerr
	}
	defer reader.Close()

	// создаем каталог для распаковки, если его нет
	if mkderr := os.MkdirAll(destDir, 0755); mkderr != nil {
		return mkderr
	}

	// итерируемся по файлам в архиве
	for _, file := range reader.File {
		// Открываем файл в архиве
		rc, rcerr := file.Open()
		if rcerr != nil {
			return rcerr
		}
		defer rc.Close()

		// создаем путь для файла в распакованном каталоге
		filePath := filepath.Join(destDir, file.Name)

		// если это директория, создаем ее
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
		} else {
			// создаем файл на диске
			dstFile, dferr := os.Create(filePath)
			if dferr != nil {
				return dferr
			}
			// устанавливаем права
			if dfcerr := dstFile.Chmod(0744); dfcerr != nil {
				return dfcerr
			}
			defer dstFile.Close()

			// инициализируем прогрессбар
			bar := progressbar.DefaultBytes(
				file.FileInfo().Size(),
				barText+" "+file.Name,
			)
			// копируем содержимое файла из архива в файл на диске
			if _, iocerr := io.Copy(io.MultiWriter(dstFile, bar), rc); iocerr != nil {
				return iocerr
			}
		}
	}

	return nil
}

// функция для копирования файлов и директорий с прогрессбаром
func CopyFiles(src, dst, barText string) error {
	// Walk проходит по всем элементам в директории src, включая вложенные директории и файлы
	return filepath.Walk(src, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			return werr
		}

		// получаем относительный путь файла относительно исходной директории
		relPath, rperr := filepath.Rel(src, path)
		if rperr != nil {
			return rperr
		}

		// создаем аналогичный путь в целевой директории
		destPath := filepath.Join(dst, relPath)

		// если элемент является директорией, создаем ее в целевой директории
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// если элемент является файлом, копируем его содержимое
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destinationFile, dferr := os.Create(destPath)
		if dferr != nil {
			return dferr
		}
		defer destinationFile.Close()

		// инициализируем прогрессбар
		bar := progressbar.DefaultBytes(
			info.Size(),
			barText+" "+info.Name(),
		)
		if _, iocerr := io.Copy(io.MultiWriter(destinationFile, bar), sourceFile); iocerr != nil {
			return iocerr
		}

		return nil
	})
}

// функция проверки наличия файлов в директории по паттерну
func FilesExistInDirectory(directory, pattern string) (bool, error) {
	// получаем список файлов в директории
	files, ferr := os.ReadDir(directory)
	if ferr != nil {
		return false, ferr
	}

	// компилируем регулярное выражение для проверки паттерна
	rx := regexp.MustCompile(pattern)

	// проверяем каждый файл в директории
	for _, file := range files {
		if rx.MatchString(file.Name()) {
			// если найден файл, соответствующий паттерну, возвращаем true
			return true, nil
		}
	}

	// если файлов с заданным паттерном не найдено, возвращаем false
	return false, nil
}
