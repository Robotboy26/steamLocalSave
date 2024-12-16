# steamLocalSave

# TODO Update
Used for saving Steam game data to locally on the hard drive.

# Purpose

Have you ever changed computer and lost steam save data because it's not being stored in the cloud. This project is your solution.

# What does it do

This project will copy your steam save game data to a zip file so that you can easily back it up or transfer it between computers.

It will save Steam game data.
Restoring Steam game data is not yet implimented>

# How to use

TODO: make a better user interface

Download the project.

```
git clone https://github.com/Robotboy26/SteamLocalSave.git
```

Enter the project directory.

```
cd SteamLocalSave/src
```

Basic command usage (Note: get help with 'python3 main.py -h').

Save your Steam game data.

```
python3 main.py -s <SteamLibrarypath>
python3 main.py -s ~/.steam/steam
```

The library path is required and 'steamapps' should be in the directory you point to but do not include it in your path.

NOT CORRECT `~/.steam/steam/steamapps'
CORRECT '~/.steam/steam'

'-s' is save, it is used to backup your game data.

Access you backed up game data.

```
cd ../SteamSaveLocal
ls
```

# How to contribute

Because of the scattered location of video game save data each game is registered seperately.
If you don't see your favorite steam games feel free to add them to the registry.
There is currently a seperate regestry for Linux and Windows installs.
The Linux regestry file is called 'SavePathDataset-Linux.txt'
