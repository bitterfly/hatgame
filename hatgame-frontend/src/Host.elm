module Host exposing (Data, default, defaultTimer, defaultWords, encode)

import Json.Encode


type alias Data =
    { players : Int
    , words : Maybe Int
    , timer : Maybe Int
    }


default : Data
default =
    { players = 2
    , words = Just defaultWords
    , timer = Just defaultTimer
    }


defaultTimer : Int
defaultTimer =
    60


defaultWords : Int
defaultWords =
    10


encode : Data -> Json.Encode.Value
encode { players, timer, words } =
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
        ]
