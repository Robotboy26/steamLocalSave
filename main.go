package main

import (
    "encoding/json"
    "fmt"
	"io"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
	"flag"
    "errors"
)

// TODO: needs improvment
type Game struct {
    Name             string
    PathList         []string `json:"pathList"`
    DeletePathList   []string `json:"deletePathList"`
    srcList          []string
    foundLocation    string
    // fix targ
    targAuto         string
    targBackup       string
}


type Config struct {
    SteamLibraryPath string `json:"steamLibraryPath"`
    LocalSaveLocation string `json:"localSaveLocation"`
    MaxBackups int `json:"maxBackups"`
    UUID string `json:"uuid"`
    Mode string `json:"mode"` // Valid modes: "save" or "restore" or "delete" or "cleanup"
    Platform string `json:"platform"`
    // Select bool `json:"select"`
	DryRun bool `json:"dryRun"`
}

// TODO: error need much improvment for the different modes
func main() {
	// Command line arguments
	var configPath = flag.String("c", "config.json", "Program settings in a json file.")
	var steamLibraryPath = flag.String("l", "", "The path to your Steam installation.")
	var localSaveLocation = flag.String("s", "SteamSaveLocal/", "The location to backup or restore from.")
	var maxBackups = flag.Int("b", 3, "The max amount of backups to make for a game")
	var UUID = flag.String("uuid", "", "Your Steam profile UUID. This is used for some steam games that store save files under a directory that is your steamUUID.")
	var mode = flag.String("m", "save", "Avalable modes are (Case does not matter): save, restore, delete, and cleanup.")
	var platform = flag.String("p", "linux", "Set this to the operating system you are using. Ex. Linux.")
	// var  = flag.Bool("select", false, "If you would like to manually select games to save, restore, delete, or cleanup select this option.")
	var dryRun = flag.Bool("dry", false, "Run the program without copying any files.")

	flag.Parse()

	fmt.Printf("%t\n", *dryRun)

	var config Config

	config.SteamLibraryPath = *steamLibraryPath
	config.LocalSaveLocation = *localSaveLocation
	config.MaxBackups = *maxBackups
	config.UUID = *UUID
	config.Mode = strings.ToLower(*mode)
	config.Platform = strings.ToLower(*platform)
	// config.Select = *select
	config.DryRun = *dryRun

    // Read the JSON config file
	// Values in a JSON file will overwrite flags
	if configFile, err := os.Open(*configPath); err == nil {
		// File exists
		data, err := io.ReadAll(configFile)
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("Error decoding config file: %v", err)
		}

	} else if errors.Is(err, os.ErrNotExist) {
		// File does not exist
	} else {
        log.Fatalf("Error reading config file: %v", err)

	}

	fmt.Printf("Config from file: %v\n", config)

	if config.SteamLibraryPath == "" {
		log.Fatal("You need to input a Steam Library Path. Use -l or set it in a config file.")
	}

	// Make Steam library path is formatted correctly
    // Ensure trailing slashes on paths
    if !strings.HasSuffix(config.SteamLibraryPath, "/") {
        config.SteamLibraryPath = fmt.Sprintf("%s/", config.SteamLibraryPath) 
    }

    if strings.HasPrefix(config.SteamLibraryPath, "~") {
        homeDir, err := os.UserHomeDir()
        if err != nil {
			fmt.Println("Failed to get Home Dir.")
            log.Fatal(err)
        }
        config.SteamLibraryPath = filepath.Join(homeDir, strings.TrimPrefix(config.SteamLibraryPath, "~"))
    }

	if !strings.HasSuffix(config.LocalSaveLocation, "/") {
        config.LocalSaveLocation = fmt.Sprintf("%s/", config.LocalSaveLocation) 
    }

    // Make sure the local library directory exists
    if _, err := os.Stat(config.LocalSaveLocation); os.IsNotExist(err) {
        err = os.Mkdir(config.LocalSaveLocation, 0755)
        if err != nil {
            log.Fatal(err)
        }
    }

	config.Mode = strings.ToLower(config.Mode)
	if config.Mode != "save" && config.Mode != "restore" && config.Mode != "delete" && config.Mode != "cleanup"{
		config.Mode = "save"
	}

    config.Platform = strings.ToLower(config.Platform)
	if config.Platform != "linux" {
		log.Fatalf("Your selected platform '%s' is not currently supported!\n", config.Platform)
	}
	// Continue

    games, err := readGamesDatabase(config.Platform)
    if err != nil {
        log.Fatal(err)
    }

    // if config.Select {
    //     games = selectGames(games)
    // }

	// Add a max threads to use in this waitGroup
    var wg sync.WaitGroup
    for _, game := range games {
        wg.Add(1)
        go func(game Game) {
            defer wg.Done()
            err := saveGame(&config, game)
            if err != nil {
                // log.Printf("Error saving game for path: '%s'. Exception: %v\n", game, err)
            } else {
				// fmt.Printf("Successfully %sd game with path: '%s'.\n", config.Mode, game.Name)
			}
        } (game)
    }
    wg.Wait()
}

func timeFormat() string {
    currentTime := time.Now()
    formattedTime := currentTime.Format("2006-01-02 15:04:05")
    return formattedTime
}

// This could likely be cleaned or split up
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
    timeCombinationBackup := fmt.Sprintf("%s-backup", timeFormat())
    targetAuto := filepath.Join(localLibrary, gameName, timeCombinationAuto, gameName)
    targetBackup := filepath.Join(localLibrary, gameName, timeCombinationBackup, gameName)
    return srcList, targetAuto, targetBackup, nil
}

func cleanupOldBackups(localLibrary string, gameName string, maxBackups int) error {
    backups, err := getAutoBackupFiles(localLibrary, gameName)
    if err != nil {
        return err
    }

	fmt.Printf("There are currently %d backups for the game '%s'\n", len(backups), gameName)

    for i := 0; i < len(backups) - maxBackups; i++ {
        oldestBackup := filepath.Join(localLibrary, gameName, backups[i])
		fmt.Printf("Removing the oldest backup: '%s'\n", oldestBackup)
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

func getBackupBackupFiles(localLibrary string, gameName string) ([]string, error) {
    files, err := os.ReadDir(filepath.Join(localLibrary, gameName))
    if err != nil {
        return nil, err
    }
    var backups []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), "backup") {
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
                return nil, fmt.Errorf("Error reading config file from game '%s' : %v", game.Name, err)
            }

            err = json.Unmarshal(gameData, &game)
            if err != nil {
                return nil, fmt.Errorf("Error decoding config file from game '%s' : %v", game.Name, err)
            }
            games = append(games, game)
        }
    }

    return games, nil
}

func findGame(steamLibrary, localLibrary string, uuid string, game Game) (Game, bool, error) {
	// Fix this by a little bit.
    var err error
    game.srcList, game.targAuto, game.targBackup, err = generatePaths(steamLibrary, localLibrary, game.Name, game.PathList)
    if err != nil {
        return game, false, err
    }

    var foundSources []string

    // Add uuid to src paths
    for _, src := range game.srcList {
        if strings.Contains(src, ";") {
            src = strings.ReplaceAll(src, ";", fmt.Sprintf("%s", uuid))
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

func saveGame(config *Config, game Game) error {
    game, foundGame, err := findGame(config.SteamLibraryPath, config.LocalSaveLocation, config.UUID, game)
    if err != nil {
        return err
    }
    if !foundGame {
        // return errors.New(fmt.Sprintf("Could not find game '%s.'", game.Name))
		return nil // Do not say anything when a game is not found
    }

	// Change to a switch statment
    if config.Mode == "save" {
        fmt.Printf("Saving game files for '%s'\n", game.Name)
        err := performCopyAndZip(game.foundLocation, game.targAuto, false)
        if err != nil {
            return err
        }
        fmt.Printf("Saved game files for '%s'\n", game.Name)
        err = cleanupOldBackups(config.LocalSaveLocation, game.Name, config.MaxBackups)
        if err != nil {
            return err
        }
        return nil
    } else if config.Mode == "restore" {
        // Create a backup first
        zipFiles, err := getAutoBackupFiles(config.LocalSaveLocation, game.Name)
		fmt.Println(zipFiles)
		if err != nil {
			return err
		}
        if len(zipFiles) == 0 {
			return errors.New("There are no saves to restore from, please run this program in save mode or rename a -backup save to -auto.")
        }
        performCopyAndZip(game.foundLocation, game.targBackup, false)
        fmt.Printf("Backed up game files for '%s'\n", game.Name)
		latestBackup := fmt.Sprintf("%s/%s.zip", filepath.Join(config.LocalSaveLocation, game.Name, zipFiles[len(zipFiles) - 1]), game.Name)
        // Make the directories to the location if they do not exists (The game could have been removed and uninstalled and just downloading it might not create the save location folders)
        err = deleteDir(game.foundLocation) // Delete the save dir after saveing but before loading it from an old save
        if err != nil {
            return err
        }
        err = unzipFile(latestBackup, game.foundLocation)
        if err != nil {
            return err
        }
        fmt.Printf("Restoring from backup '%s' to game files\n", latestBackup)
        return nil
    } else if config.Mode == "delete" {
		// log.Fatal("delete is not avalible at this time.")
        // TODO: add saftey to this
        // Create a backup first
        performCopyAndZip(game.foundLocation, game.targBackup, false)
        fmt.Printf("Backed up game files for '%s'\n", game.Name)
        err = deleteDir(game.targBackup)
        if err != nil {
            return err
        }
		var deletePaths []string
        if len(game.DeletePathList) == 0 {
			deletePaths = append(deletePaths, game.foundLocation)
            // return errors.New("There is no delete path, please edit to database file to fix this.")
        } else {
			deletePaths = game.DeletePathList
		}
        // We already know that game.foundLocation exists and is a dir
        // But I also want support for deleting from multiple paths
        var deletedSomething bool
        for x := 0; x < len(deletePaths); x ++ {
			var pathToDelete string
			if strings.HasPrefix(deletePaths[x], "~") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				pathToDelete = fmt.Sprintf("%s%s", homeDir, strings.TrimPrefix(deletePaths[x], "~"))
			} else {
				pathToDelete = deletePaths[x]
			}
			fmt.Printf("Starting to delete path '%s'\n", pathToDelete)
            info, err := os.Stat(pathToDelete)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println("Save data path does not exist!")
                continue
            } else if err != nil {
                return err
            }
            if info.IsDir() {
				// Step back from the save data location one directory at a time to delete the farthest back path that will still not delete any unwanted data.
                var dirPath string
				dirPath = pathToDelete
				// Step back one dir at a time
                for {
					for i := 1; i < len(dirPath); i++ {
                        if dirPath[len(dirPath) - i] == byte("/"[0]) {
                            dirPath = dirPath[0: len(dirPath) - i]
                            break
                        }
                    }
                    // Check how many dirs are in the path to delete
                    entries, err := os.ReadDir(dirPath)
                    if err != nil {
                        return err
                    }

                    dirCount := 0
                    for _, entry := range entries {
                        if entry.IsDir() {
                            dirCount++
                        }
                    }
                    if dirCount < 2 {
						// Only delete that path if there is only one or less subdirectorys
						fmt.Printf("Deleted dir '%s'\n", dirPath)
                        deleteDir(dirPath)
                        deletedSomething = true
                    } else {
						fmt.Printf("Dir '%s' has more than 1 subdirectory\n", dirPath)
						fmt.Printf("Deleted dir '%s'\n", pathToDelete)
						deleteDir(pathToDelete)
						break
					}
                }
            }
        }
        if deletedSomething {
            return nil
        } else {
            return errors.New("There was some kind of problem during the delete function that caused nothing to be delete. This is weird because it should have errored somewhere else.")
        }
    } else if config.Mode == "cleanup" {
		//       autoBackups, err := getAutoBackupFiles(config.LocalSaveLocation, game.Name)
		// if err != nil {
		// 	return err
		// }
		// latestAutoBackup := autoBackups[len(autoBackups) - 1]
		// for i := 0; i < len(autoBackups); i++ {
		// 	if autoBackups[i] != latestAutoBackup {
		// 		fmt.Printf("Deleted '%s'\n", autoBackups[i])
		// 		deleteDir(filepath.Join(config.LocalSaveLocation, game.Name, autoBackups[i]))
		// 	}
		// }

		// Keep max backups auto backups
        err = cleanupOldBackups(config.LocalSaveLocation, game.Name, config.MaxBackups)
        if err != nil {
            return err
        }

		backupBackups, err := getBackupBackupFiles(config.LocalSaveLocation, game.Name)
		if err != nil {
			return err
		}
		latestBackupBackup := backupBackups[len(backupBackups) - 1]
		for i := 0; i < len(backupBackups); i++ {
			if backupBackups[i] != latestBackupBackup {
				fmt.Printf("Deleted '%s'\n", backupBackups[i])
				deleteDir(filepath.Join(config.LocalSaveLocation, game.Name, backupBackups[i]))
			}
		}
		return nil
	}
    return errors.New("Option error, no operation was ran.")
}
