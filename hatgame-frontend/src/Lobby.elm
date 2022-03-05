module Lobby exposing (Data, ProcessState(..))

import Containers.Game


type ProcessState
    = WaitingPlayers
    | ReadyToStart


type alias Data =
    { game : Containers.Game.Game
    , processState : ProcessState
    }
