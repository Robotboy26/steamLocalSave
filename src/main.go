package main

import (
    // "bufio"
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
    Name             string
    PathList         []string `json:"pathList"`
    srcList          []string
    foundLocation    string
    // fix targ
    targAuto         string
    targBackup       string
    DeletePaths      []string `json:"deleteList"`
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

func generatePaths(steamLibrary, localLibrary, gameName string, savePaths []string) ([]string, string, string, error) {
    var srcList []string
    for _, path := range savePaths {
        if !strings.Contains(path, "~") {
            src := filepath.Join(steamLibrary, path)
            srcList = append(srcList, src)
        } else {
            src, err := os.UserHomeDir()
            if err != nil {
                return nil, "", "", err
            }
            src = filepath.Join(src, strings.TrimPrefix(path, "~"))
            srcList = append(srcList, src)
        }
    }
    timeCombinationAuto := fmt.Sprintf("%s-auto", timeFormat())
    timeCombinationBackup := fmt.Sprintf("%s-auto", timeFormat())
    targetAuto := filepath.Join(localLibrary, gameName, timeCombinationAuto, gameName)
    targetBackup := filepath.Join(localLibrary, gameName, timeCombinationBackup, gameName)
    return srcList, targetAuto, targetBackup, nil
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

func readGamesDatabase(platform string) ([]Game, error) {
    dbDir := fmt.Sprintf("../database/%s", platform)
    var games []Game

    // Read the database dir
    dbFiles, err := os.ReadDir(dbDir)
    if os.IsNotExist(err) {
        return nil, errors.New("The database dir is missing or your platform is not supported.")
    } else if err != nil {
        return nil, err
    }

    for _, file := range dbFiles {
        if file.IsDir() {
            continue
        }

        if strings.HasSuffix(file.Name(), ".json") {
            var game Game

            gameName := strings.TrimSuffix(file.Name(), ".json")
            game.Name = gameName
            filePath := filepath.Join(dbDir, file.Name())
            gameData, err := os.ReadFile(filePath)
            if err != nil {
                return nil, fmt.Errorf("Error reading config file: %v", err)
            }

            err = json.Unmarshal(gameData, &game)
            if err != nil {
                return nil, fmt.Errorf("Error decoding config file: %v", err)
            }
            games = append(games, game)
        }
    }

    return games, nil
}

func findGame(steamLibrary, localLibrary string, uuid int, game Game) (Game, bool, error) {
    var err error
    game.srcList, game.targAuto, game.targBackup, err = generatePaths(steamLibrary, localLibrary, game.Name, game.PathList)
    if err != nil {
        return game, false, err
    }

    var foundSources []string

    // Add uuid to src paths
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

func saveGame(steamLibrary, localLibrary, option string, maxBackups, uuid int, game Game) (bool, error) {
    game, foundGame, err := findGame(steamLibrary, localLibrary, uuid, game)
    if err != nil {
        return false, err
    }
    if !foundGame {
        return false, nil
    }

    if option == "save" {
        fmt.Printf("Saving game files for '%s'\n", game.Name)
        err := performCopy(game.foundLocation, game.targAuto, false)
        if err != nil {
            return false, err
        }
        err = deleteDir(game.targAuto)
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
        performCopy(game.foundLocation, game.targBackup, false)
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
    } else if option == "delete" {
        // TODO: add saftey to this
        // Create a backup first
        performCopy(game.foundLocation, game.targAuto, false)
        for x := 0; x < len(game.DeletePaths) - 1; x ++ {
            pathToDelete := game.DeletePaths[x]
            info, err := os.Stat(pathToDelete)
            if os.IsNotExist(err) {
                continue
            } else if err != nil {
                return false, err
            }
            if info.IsDir() {
                deleteDir(pathToDelete)
            }
        }
    }
    optionError := errors.New("Option error, no operation was ran.")
    return false, optionError
}

func saveGames(config *Config) {
    games, err := readGamesDatabase(config.Platform)
    if err != nil {
        log.Fatal(err)
    }

    var wg sync.WaitGroup
    for _, game := range games {
        wg.Add(1)
        go func(game Game) {
            defer wg.Done()
            // TODO: when changing this to use config use a copy of config not a pointer to it as too not accidently change a global value in a unwanted way
            status, err := saveGame(config.SteamLibraryPath, config.LocalLibrary, config.Mode, config.MaxBackups, config.UUID, game)
            if err != nil {
                log.Printf("Error saving game for path: %s. Exception: %v\n", game, err)
            }
            if status {
                fmt.Printf("Successfully saved game with path: %s.\n", game.Name)
            }
        }(game)
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
