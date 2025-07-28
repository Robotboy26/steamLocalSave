package main

import (
    "github.com/nsf/termbox-go"
)

func selectGames(games []Game) []Game {
    var selectedGames []Game
    if err := termbox.Init(); err != nil {
        panic(err)
    }
    defer termbox.Close()

    selected := make(map[int]bool)
    cursor := 0

    for {
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        for i, game := range games {
            if selected[i] {
                printColor(i, " [x] " + game.Name)
            } else {
                printColor(i, " [ ] " + game.Name)
            }
        }
        printCursor(cursor)
        termbox.Flush()
        ev := termbox.PollEvent()
        switch ev.Type {
        case termbox.EventKey:
            switch ev.Key {
            case termbox.KeyArrowDown:
                if cursor < len(games)-1 {
                    cursor++
                }
            case termbox.KeyArrowUp:
                if cursor > 0 {
                    cursor--
                }
            case termbox.KeySpace:
                selected[cursor] = !selected[cursor]
            case termbox.KeyEnter:
                for i, isSelected := range selected {
                    if isSelected {
                        selectedGames = append(selectedGames, games[i])
                    }
                }
                return selectedGames
            case termbox.KeyCtrlC:
                return nil
            }
            switch ev.Ch {
                case 106: // j
                if cursor < len(games)-1 {
                    cursor++
                }
                case 107: 
                if cursor > 0 {
                    cursor--
                }
            }
        }
    }
}

func printCursor(cursor int) {
    termbox.SetCell(0, cursor, '>', termbox.ColorWhite, termbox.ColorDefault)
}

func printColor(i int, text string) {
    x, y := 2, i
    for _, ch := range text {
        termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorDefault)
        x++
    }
}
