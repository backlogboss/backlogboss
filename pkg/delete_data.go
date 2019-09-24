package pkg

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

func DeleteData() {
	prompt := promptui.Prompt{
		Label:     "Delete all settings and data",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		os.Exit(1)
	}

	if result == "y" {
		err := removeConfDir()
		if err != nil {
			fmt.Printf("error deleting %v", err)
			os.Exit(1)
		}
		fmt.Println("Deleted.")
		return
	}
}

func removeConfDir() error {
	err := os.RemoveAll(getConfDir())
	if err != nil {
		return err
	}
	return nil
}
