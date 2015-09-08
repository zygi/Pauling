package main

import (
	"sync"

	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/PlayerStatsScraper/steamid"
	rconwrapper "github.com/TF2Stadium/TF2RconWrapper"
)

type Pauling int
type Noreply struct{}

func (_ *Pauling) VerifyInfo(info *models.ServerRecord, nop *Noreply) error {
	_, err := rconwrapper.NewTF2RconConnection(info.Host, info.RconPassword)

	return err
}

func (_ *Pauling) SetupVerifier(args *models.ServerBootstrap, nop *Noreply) error {
	s := NewServer()
	LobbyMutexMap[args.LobbyId].Lock()
	LobbyServerMap[args.LobbyId] = s
	LobbyMutexMap[args.LobbyId].Unlock()
	s.LobbyId = args.LobbyId
	s.Info = args.Info
	for _, playerId := range args.BannedPlayers {
		s.AllowedPlayers[playerId] = false
	}
	for _, playerId := range args.Players {
		s.AllowedPlayers[playerId] = true
	}
	s.StartVerifier()
	go s.CommandListener()

	return nil
}

func (_ *Pauling) SetupServer(args *models.Args, nop *Noreply) error {
	s := NewServer()
	s.LobbyId = args.Id
	s.Info = args.Info
	s.Type = args.Type
	s.League = args.League
	s.Map = args.Map

	err := s.Setup()
	if err != nil {
		Logger.Warning(err.Error())
		return err
	}

	err = s.StartVerifier()
	if err != nil {
		Logger.Warning(err.Error())
		return err
	}

	LobbyServerMap[s.LobbyId] = s
	LobbyMutexMap[s.LobbyId] = &sync.Mutex{}
	return nil
}

func (_ *Pauling) End(args *models.Args, nop *Noreply) error {
	s := LobbyServerMap[args.Id]
	s.End()
	return nil
}

func (_ *Pauling) AllowPlayer(args *models.Args, nop *Noreply) error {
	commId, _ := steamid.SteamIdToCommId(args.SteamId)
	LobbyServerMap[args.Id].AllowedPlayers[commId] = true
	return nil
}

func (_ *Pauling) DisallowPlayer(args *models.Args, nop *Noreply) error {
	s := LobbyServerMap[args.Id]
	commId, _ := steamid.SteamIdToCommId(args.SteamId)
	if s.IsPlayerAllowed(commId) {
		delete(s.AllowedPlayers, commId)
	}
	return nil
}

func (_ *Pauling) SubstitutePlayer(args *models.Args, nop *Noreply) error {
	s := LobbyServerMap[args.Id]
	commId, _ := steamid.SteamIdToCommId(args.SteamId)
	commId2, _ := steamid.SteamIdToCommId(args.SteamId2)
	s.Substitutes[commId2] = commId
	return nil
}

func (_ *Pauling) GetEvent(args *models.Args, event *models.Event) error {
	e := PopEvent()
	*event = e
	return nil
}
