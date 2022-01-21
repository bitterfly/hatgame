module Home exposing (Data)

import Containers.Statistics


type alias Data =
    { gameId : Maybe Int
    , stats : Maybe Containers.Statistics.Statistics
    }
