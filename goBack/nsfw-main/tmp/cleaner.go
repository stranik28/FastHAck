package tmp

import (
	"fmt"
	"os"
	"path/filepath"
)

func ClearDir() error {
	dir := "framess"
	// Получаем список всех файлов и папок в указанной директории
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("не удалось прочитать директорию: %v", err)
	}

	// Удаляем каждый файл и папку
	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())
		err = os.RemoveAll(entryPath)
		if err != nil {
			return fmt.Errorf("не удалось удалить %s: %v", entryPath, err)
		}
	}

	return nil
}
