package main

import (
    "os"
    "io"
    "path/filepath"
    "fmt"
)

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
        return fmt.Errorf("failed to delete directory %s: %v", path, err)
    }
    return nil
}
