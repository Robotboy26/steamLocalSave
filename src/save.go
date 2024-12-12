package main

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func log(message string) {
    fmt.Println(message)
}

func timeFormat() string {
    currentTime := time.Now()
    formattedTime := currentTime.Format("2006-01-02 15:04:05")
    return formattedTime
}

func copyDir(src string, dst string) error {
    srcFi, err := os.Stat(src)
    if err != nil {
        return err
    }

    err = os.MkdirAll(dst, srcFi.Mode())
    if err != nil {
        return err
    }

    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())

        if entry.IsDir() {
            err = copyDir(srcPath, dstPath)
            if err != nil {
                return err
            }
        } else {
            err = copyFile(srcPath, dstPath)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func copyFile(src string, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    if err != nil {
        return err
    }

    return out.Close()
}

func createZipArchive(src string, dst string) error {
    zipFile, err := os.Create(fmt.Sprintf("%s.createZipArchive", dst))
    if err != nil {
        return err
    }
    defer zipFile.Close()

    archive := zip.NewWriter(zipFile)
    defer archive.Close()

    err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        header, err := zip.FileInfoHeader(info)
        if err != nil {
            return err
        }

        header.Name = relPath

        writer, err := archive.CreateHeader(header)
        if err != nil {
            return err
        }

        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()

        _, err = io.Copy(writer, file)
        return err
    })
    return err
}

func saveGame(SteamLibrary string, LocalLibrary string, maxBackups int, option string, path string) {
    parts := strings.Split(path, "|")
    if len(parts) != 3 {
        return
    }

    gameName, savePath, backupPath := parts[0], parts[1], parts[2]
    src, targ := generatePaths(SteamLibrary, gameName, savePath, backupPath, LocalLibrary)

    if _, err := os.Stat(src); os.IsNotExist(err) {
        return
    }

    if _, err := os.Stat(fmt.Sprintf("%s%s", LocalLibrary, gameName)); os.IsNotExist(err) {
        performCopy(src, targ, gameName, false)
        log(fmt.Sprintf("Saved data for %s", gameName))
    } else {
        if strings.ToLower(option) == "all" {
            entries, err := os.ReadDir(fmt.Sprintf("%s%s", LocalLibrary, gameName))
            if err != nil {
                log(fmt.Sprintf("Error reading directory: %v", err))
                return
            }
            zipFiles := []string{}
            for _, file := range entries {
                if strings.HasSuffix(file.Name(), ".createZipArchive") {
                    zipFiles = append(zipFiles, file.Name())
                }
            }
            // sort zipFiles
            if len(zipFiles) > maxBackups {
                log(fmt.Sprintf("You have %d backups for game: '%s'", len(zipFiles), gameName))
                for len(zipFiles) > maxBackups {
                    log(fmt.Sprintf("More than %d backups for game '%s'. Removing the oldest: %s", maxBackups, gameName, zipFiles[len(zipFiles)-1]))
                    oldestBackup := filepath.Join(fmt.Sprintf("%s%s", LocalLibrary, gameName), zipFiles[len(zipFiles)-1])
                    err := os.Remove(oldestBackup)
                    if err != nil {
                        log(fmt.Sprintf("Error removing file: %v", err))
                        return
                    }
                }
                performCopy(src, targ, gameName, false)
                log(fmt.Sprintf("Saved data for %s", gameName))
            }
        }
    }
}

func saveGames(SteamLibrary string, LocalLibrary string, maxBackups int, option string) {
    file, err := os.Open("../SavePathDatasetLinux.txt")
    if err != nil {
        log(fmt.Sprintf("Error opening file: %v", err))
        return
    }
    defer file.Close()

    // SavePaths := []string{}
}

func generatePaths(SteamLibrary string, gameName string, savePath string, backupPath string, LocalLibrary string) (string, string) {
    var src string
    if !strings.Contains(savePath, "~") {
        src = filepath.FromSlash(fmt.Sprintf("%s/%s", SteamLibrary, savePath))
    } else {
        src = filepath.FromSlash(savePath)
        src, _ = filepath.Abs(src)
    }

    targ := filepath.FromSlash(fmt.Sprintf("%s%s/%s/%s", LocalLibrary, gameName, timeFormat(), backupPath))
    return src, targ
}

func performCopy(src string, targ string, gameName string, dryRun bool) {
    if !dryRun {
        log(fmt.Sprintf("Saving files for '%s'", gameName))
        err := os.MkdirAll(targ, os.ModePerm)
        if err != nil {
            log(fmt.Sprintf("Error creating directory: %v", err))
            return
        }
        err = copyDir(src, targ)
        if err != nil {
            log(fmt.Sprintf("Error copying directory: %v", err))
            return
        }
        log(fmt.Sprintf("Creating zip archive of '%s'", gameName))
        zipPath := targ
        err = createZipArchive(zipPath, targ)
        if err != nil {
            log(fmt.Sprintf("Error zipping directory: %v", err))
            return
        }
        os.RemoveAll(targ)
    }
}

func main() {
    SteamLibrary := "/media/robot/steamgames/SteamLibrary"
    maxBackups := 2
    option := ""
    if len(os.Args) > 2 {
        option = os.Args[1]
        SteamLibrary = os.Args[2]
    }
    if SteamLibrary == "" {
        log("Please provide SteamLibrary path")
        return
    }

    SteamLibrary = filepath.FromSlash(fmt.Sprintf("%ssteamapps/", SteamLibrary))

    LocalLibrary := ""
    if len(os.Args) > 3 {
        LocalLibrary = os.Args[3]
    }
    if LocalLibrary == "" {
        LocalLibrary = "../SteamSaveLocal/"
    }

    if _, err := os.Stat(LocalLibrary); os.IsNotExist(err) {
        err := os.Mkdir(LocalLibrary, os.ModePerm)
        if err != nil {
            log(fmt.Sprintf("Error creating directory: %v", err))
            return
        }
    }

    saveGames(SteamLibrary, LocalLibrary, maxBackups, option)
}
