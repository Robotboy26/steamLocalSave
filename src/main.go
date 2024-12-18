package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    "errors"
)

// TODO: needs improvment
type Game struct {
    Name string
    libraryPathList []string
    constantPathList []string
    srcList []string
    foundLocation string
    targ string
}

var debugMode bool

func logDebug(message string) {
    if debugMode {
        fmt.Println(message)
    }
}

func timeFormat() string {
    currentTime := time.Now()
    formattedTime := currentTime.Format("2006-01-02 15:04:05")
    return formattedTime
}

func generatePaths(steamLibrary, localLibrary, gameName, suffix string, savePaths []string) ([]string, string, error) {
    var srcList []string
    for _, path := range savePaths {
        if !strings.Contains(path, "~") {
            src := filepath.Join(steamLibrary, path)
            srcList = append(srcList, src)
        } else {
            src, err := os.UserHomeDir()
            if err != nil {
                return srcList, "", err
            }
            src = filepath.Join(src, strings.TrimPrefix(path, "~"))
            srcList = append(srcList, src)
        }
    }
    timeCombination := fmt.Sprintf("%s-%s", timeFormat(), suffix)
    target := filepath.Join(localLibrary, gameName, timeCombination, gameName)
    return srcList, target, nil
}

func performCopy(src string, targ string, dryRun bool) error {
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

        return nil
    }
    return nil
}

func cleanupOldBackups(localLibrary string, gameName string, maxBackups int) error {
    backups, err := getAutoBackupFiles(localLibrary, gameName)
    if err != nil {
        return err
    }

    for i := 0; i < len(backups)-maxBackups; i++ {
        oldestBackup := filepath.Join(localLibrary, gameName, backups[i])
        logDebug(fmt.Sprintf("Removing the oldest backup: %s", oldestBackup))
        if err := os.RemoveAll(oldestBackup); err != nil {
            return err
        }
    }
    return nil
}

func getAutoBackupFiles(localLibrary string, gameName string) ([]string, error) {
    files, err := os.ReadDir(filepath.Join(localLibrary, gameName))
    if err != nil {
        return nil, err
    }
    var backups []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), "auto") {
            backups = append(backups, file.Name())
        }
    }
    return backups, nil
}

func readGamesDatabase(platform string) ([]string, error) {
    // TODO
    // return []game
    // read database dir
    // then platform dir
    // then all games to create a list of game structs with the needed information
    fileName := fmt.Sprintf("../SavePathDataset-%s.txt", platform)
    file, err := os.Open(fileName)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var savePaths []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "#") || line == "" {
            continue
        }
        lineParts := strings.SplitN(line, "#", 2)
        savePaths = append(savePaths, strings.TrimSpace(lineParts[0]))
    }
    return savePaths, scanner.Err()
}

func findGame(steamLibrary, localLibrary, path string, uuid int) (Game, bool, error) {
    var game = Game{}

    parts := strings.SplitN(path, "|", 2)
    game.Name = parts[0]
    game.constantPathList = strings.Split(parts[1], "|")

    var err error
    game.srcList, game.targ, err = generatePaths(steamLibrary, localLibrary, game.Name, "auto", game.constantPathList)
    if err != nil {
        return game, false, err
    }

    var foundSources []string

    for _, src := range game.srcList {
        if uuid != 0 && strings.Contains(src, ";") {
            src = strings.ReplaceAll(src, ";", fmt.Sprintf("%d", uuid))
        }
        _, err := os.Stat(src)
        if os.IsNotExist(err) {
            return game, false, nil
        } else if err != nil {
            return game, false, err
        }
        foundSources = append(foundSources, src)
    }

    if len(foundSources) == 1 {
        game.foundLocation = foundSources[0]
        return game, true, nil
    } else {
        return game, false, nil
    }
}

func saveGame(steamLibrary, localLibrary, option, path string, maxBackups, uuid int) (bool, error) {
    game, foundGame, err := findGame(steamLibrary, localLibrary, path, uuid)
    if err != nil {
        return false, err
    }
    if !foundGame {
        return false, nil
    }

    if option == "save" {
        fmt.Printf("Saving game files for '%s'\n", game.Name)
        err := performCopy(game.foundLocation, game.targ, false)
        if err != nil {
            return false, err
        }
        err = deleteDir(game.targ)
        if err != nil {
            return false, err
        }
        err = cleanupOldBackups(localLibrary, game.Name, maxBackups)
        if err != nil {
            return false, err
        }
        return true, nil
    } else if option == "restore" {
        // Create a backup first
        performCopy(game.foundLocation, game.targ, false)
        zipFiles, err := getAutoBackupFiles(localLibrary, game.Name)
        if err != nil || len(zipFiles) == 0 {
            return true, nil
        }
        latestBackup := filepath.Join(localLibrary, game.Name, zipFiles[len(zipFiles)-1])
        logDebug(fmt.Sprintf("Restoring from backup '%s' to game files", latestBackup))
        err = unzipFile(latestBackup, game.foundLocation)
        if err != nil {
            return false, err
        }
        return true, nil
    }
    optionError := errors.New("Option error, no operation was ran.")
    return false, optionError
}

func saveGames(config *Config) {
    savePaths, err := readGamesDatabase(config.Platform)
    if err != nil {
        log.Fatalf("Unable to read save paths: %v", err)
    }

    var wg sync.WaitGroup
    for _, path := range savePaths {
        wg.Add(1)
        go func(pathToGame string) {
            defer wg.Done()
            // TODO: when changing this to use config use a copy of config not a pointer to it as too not accidently change a global value in a unwanted way
            status, err := saveGame(config.SteamLibraryPath, config.LocalLibrary, config.Mode, pathToGame, config.MaxBackups, config.UUID)
            if err != nil {
                log.Printf("Error saving game for path: %s. Exception: %v\n", pathToGame, err)
            }
            if status {
                fmt.Printf("Successfully saved game with path: %s.\n", pathToGame)
            }
        }(path)
    }
    wg.Wait()
}

type Config struct {
    SteamLibraryPath string `json:"steamLibraryPath"`
    LocalLibrary     string `json:"localLibrary"`
    MaxBackups       int    `json:"maxBackups"`
    UUID             int    `json:"uuid"`
    Mode             string `json:"mode"` // "save" or "restore"
    DebugMode        bool   `json:"debugMode"`
    Platform         string `json:"platform"`
}

func main() {
    // Read the JSON config file
    configFile := "config.json"
    data, err := os.ReadFile(configFile)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        log.Fatalf("Error decoding config file: %v", err)
    }

    debugMode = config.DebugMode
    if debugMode {
        logDebug(fmt.Sprintf("Configuration: %+v", config))
    }

    // Ensure trailing slashes on paths
    if !strings.HasSuffix(config.SteamLibraryPath, "/") {
        config.SteamLibraryPath += "/"
    }
    if strings.HasPrefix(config.SteamLibraryPath, "~") {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            log.Fatal(err)
        }
        config.SteamLibraryPath = filepath.Join(homeDir, strings.TrimPrefix(config.SteamLibraryPath, "~"))
    }

    if config.LocalLibrary == "" {
        config.LocalLibrary = "../SteamSaveLocal/"
    } else if !strings.HasSuffix(config.LocalLibrary, "/") {
        config.LocalLibrary += "/"
    }

    // Make sure the local library directory exists
    if _, err := os.Stat(config.LocalLibrary); os.IsNotExist(err) {
        err = os.Mkdir(config.LocalLibrary, 0755)
        if err != nil {
            log.Fatal(err)
        }
    }

    config.Platform = strings.ToLower(config.Platform)

    // Start saving or restoring games
    saveGames(&config)
}
