module Words exposing (Data, ProcessState(..))

import Containers.Game


type ProcessState
    = Typing
    | Done


type alias Data =
    { game : Containers.Game.Game
    , words : List String
    , currentWord : Maybe String
    , processState : ProcessState
    }
