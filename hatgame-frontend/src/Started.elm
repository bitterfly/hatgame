module Started exposing (Data, ProcessState(..))

import Containers.Game
import Containers.User


type ProcessState
    = StorytellerWaiting
    | StorytellerActive
    | NotStoryteller (Maybe Containers.User.User)


type alias Data =
    { game : Containers.Game.Game
    , currentWord : Maybe String
    , partner : Maybe Containers.User.User
    , timer : Maybe Int
    , processState : ProcessState
    }
