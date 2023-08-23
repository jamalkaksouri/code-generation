package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	VERSION = "1.1.0"
)

func putAsciiArt(s string) {
	for _, c := range s {
		d := string(c)
		switch string(c) {
		case "@":
			color.Set(color.BgGreen)
			d = " "
		case "!":
			color.Set(color.BgGreen)
			d = " "
		case ":":
			color.Set(color.BgCyan)
			d = " "
		case " ":
			color.Unset()
			d = " "
		case "\n":
			color.Unset()
		}
		fmt.Print(d)
	}
	color.Unset()
}

func printUpdateName() {
	nameClr := color.New(color.FgHiWhite)
	txt := nameClr.Sprintf("        .: A tool for generating unique, randomized codes :.")
	fmt.Fprintf(color.Output, "%s", txt)
}

func printOneliner() {
	handleClr := color.New(color.FgHiBlue)
	versionClr := color.New(color.FgGreen)
	textClr := color.New(color.FgHiBlack)
	spc := strings.Repeat(" ", 10-len(VERSION))
	txt := textClr.Sprintf("        by jamal kaksouri (") + handleClr.Sprintf("@jamalkaksouri") + textClr.Sprintf(")") + spc + textClr.Sprintf("version ") + versionClr.Sprintf("%s", VERSION)
	fmt.Fprintf(color.Output, "%s", txt)
}

func Banner() {
	fmt.Println()

	putAsciiArt(" @@@@@@@   @@@@@@   @@@@@@@   @@@@@@@@   @@@@@@@@  @@@@@@@@  @@@  @@@  \n")
	putAsciiArt("@@@@@@@@  @@@@@@@@  @@@@@@@@  @@@@@@@@  @@@@@@@@@  @@@@@@@@  @@@@ @@@  \n")
	putAsciiArt("!@@       @@!  @@@  @@!  @@@  @@!       !@@        @@!       @@!@!@@@  \n")
	putAsciiArt("!@!       !@!  @!@  !@!  @!@  !@!       !@!        !@!       !@!!@!@!  \n")
	putAsciiArt("!@!       @!@  !@!  @!@  !@!  @!!!:!    !@! @!@!@  @!!!:!    @!@ !!@!  \n")
	putAsciiArt("!!!       !@!  !!!  !@!  !!!  !!!!!:    !!! !!@!!  !!!!!:    !@!  !!!  \n")
	putAsciiArt(":!!       !!:  !!!  !!:  !!!  !!:       :!!   !!:  !!:       !!:  !!!  \n")
	putAsciiArt(":!:       :!:  !:!  :!:  !:!  :!:       :!:   !::  :!:       :!:  !:!  \n")
	putAsciiArt(" ::: :::  ::::: ::   :::: ::   :: ::::   ::: ::::   :: ::::   ::   ::  \n")
	putAsciiArt(" :: :: :   : :  :   :: :  :   : :: ::    :: :: :   : :: ::   ::    :   \n")
	fmt.Println()
	printUpdateName()
	// fmt.Println()
	fmt.Println()
	printOneliner()
	fmt.Println()
}
