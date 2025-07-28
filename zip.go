package main

import (
    "os"
    "io"
    "archive/zip"
    "path/filepath"
)

func createZip(zipPath string) error {
    zipFile, err := os.Create(zipPath + ".zip")
    if err != nil {
        return err
    }
    defer zipFile.Close()

    zipWriter := zip.NewWriter(zipFile)
    defer zipWriter.Close()

    return filepath.Walk(zipPath, func(file string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(zipPath, file)
        if err != nil {
            return err
        }

        zipFileWriter, err := zipWriter.Create(relPath)
        if err != nil {
            return err
        }

        fileReader, err := os.Open(file)
        if err != nil {
            return err
        }
        
        // Ensure fileReader is closed after use
        defer func() {
            _ = fileReader.Close()
        }()

        _, err = io.Copy(zipFileWriter, fileReader)
        if err != nil {
            return err
        }

        return nil // Return nil to continue walking
    })
}

func unzipFile(zipPath string, extractTo string) error {
    r, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        fpath := filepath.Join(extractTo, f.Name)
        if f.FileInfo().IsDir() {
            if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
                return err
            }
            continue
        }

        // Ensure the directory for the file exists
        if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        // Create the destination file
        dstFile, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }

        // Ensure srcFile is closed after use
        srcFile, err := f.Open()
        if err != nil {
            _ = dstFile.Close() // Ensure to close dstFile
            return err
        }

        // Conduct the copy operation
        _, err = io.Copy(dstFile, srcFile)
        if err != nil {
            _ = dstFile.Close() // Ensure to close dstFile
            _ = srcFile.Close() // Ensure to close srcFile
            return err
        }

        // Close both the source and destination files
        if err := dstFile.Close(); err != nil {
            return err
        }
        if err := srcFile.Close(); err != nil {
            return err
        }
    }
    return nil
}
