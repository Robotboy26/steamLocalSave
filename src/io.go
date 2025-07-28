package main

import (
    "os"
    "io"
    "path/filepath"
	"log"
)

func performCopyAndZip(src string, targ string, dryRun bool) error {
    if !dryRun {
        err := os.MkdirAll(filepath.Dir(targ), 0755)
        if err != nil {
            return err
        }
        err = copyDir(src, targ)
        if err != nil {
            return err
        }
        err = createZip(targ)
        if err != nil {
            return err
        }
		// Delete the folder once the zip file is created by performCopy.
        err = deleteDir(targ)
        if err != nil {
            return err
        }
        return nil
    }
    return nil
}

func copyFile(srcFile string, dstFile string) error {
    src, err := os.Open(srcFile)
    if err != nil {
        return err
    }
    defer src.Close()

    dst, err := os.Create(dstFile)
    if err != nil {
        return err
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    return err
}

func copyDir(src string, dst string) error {
    return filepath.Walk(src, func(file string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(src, file)
        if err != nil {
            return err
        }

        dstFile := filepath.Join(dst, relPath)

        if info.IsDir() {
            return os.MkdirAll(dstFile, os.ModePerm)
        }

        return copyFile(file, dstFile)
    })
}

func deleteDir(path string) error {
    if err := os.RemoveAll(path); err != nil {
		log.Fatalf("failed to delete directory %s: %v", path, err)
        return err
    }
    return nil
}
