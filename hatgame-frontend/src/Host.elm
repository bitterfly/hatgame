module Host exposing (Data, default, defaultTimer, defaultWords, encode)

import Json.Encode


type alias Data =
    { players : Int
    , words : Maybe Int
    , stages : Int
    , timer : Maybe Int
    , maxStages : Int
    }


default : Data
default =
    { players = defaultPlayers
    , words = Just defaultWords
    , timer = Just defaultTimer
    , stages = defaultStages
    , maxStages = defaultMaxStages
    }


defaultPlayers : Int
defaultPlayers =
    2


defaultTimer : Int
defaultTimer =
    60


defaultWords : Int
defaultWords =
    10


defaultStages : Int
defaultStages =
    1


defaultMaxStages : Int
defaultMaxStages =
    3


encode : Data -> Json.Encode.Value
encode { players, timer, words, stages } =
    let
        defaultedTimer =
            case timer of
                Nothing ->
                    defaultTimer

                Just n ->
                    n
    in
    let
        defaultedWords =
            case words of
                Nothing ->
                    defaultWords

                Just n ->
                    n
    in
    Json.Encode.object
        [ ( "Players", Json.Encode.int players )
        , ( "Timer", Json.Encode.int defaultedTimer )
        , ( "Words", Json.Encode.int defaultedWords )
        , ( "Stages", Json.Encode.int stages )
        ]
