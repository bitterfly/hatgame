module Containers.Game exposing
    ( Game
    , Result
    , Team
    , decode
    , decodeTeam
    , default
    , getUsername
    , playerById
    , result
    , show
    , showResult
    , showTeams
    )

import Containers.User exposing (User, decodeUser)
import Html exposing (Html, br, div, text)
import Json.Decode exposing (Decoder)
import Json.Encode


type alias Game =
    { id : Int
    , numPlayers : Int
    , timer : Int
    , numWords : Int
    , numStages : Int
    , host : Int
    , players : List User
    }


getUsername : List User -> Int -> Maybe String
getUsername us id =
    case us of
        [] ->
            Nothing

        t :: ts ->
            if t.id == id then
                Just t.username

            else
                getUsername ts id


default : Game
default =
    { id = 1
    , numPlayers = 4
    , timer = 60
    , numWords = 10
    , numStages = 1
    , host = 1
    , players = [ { id = 1, email = "foo", username = "username" } ]
    }


type alias Team =
    { playerOne : Int
    , playerTwo : Int
    , score : Int
    }


type Result
    = Win
    | Lose
    | Tie


showTeam : Team -> String
showTeam { playerOne, playerTwo, score } =
    String.join " " [ "[", String.fromInt playerOne, ",", String.fromInt playerTwo, "Score:", String.fromInt score, "]" ]


showResult : Result -> String
showResult res =
    case res of
        Win ->
            "You win! :)"

        Lose ->
            "You lose. :("

        Tie ->
            "It's a tie. :|"


result : Containers.User.User -> List Team -> Result
result u ts =
    case ts of
        [] ->
            Lose

        t :: tts ->
            goResult u (tieOrWin ts) t.score (t :: tts)


tieOrWin : List Team -> Result
tieOrWin ts =
    case ts of
        t :: s :: _ ->
            if t.score == s.score then
                Tie

            else
                Win

        _ :: [] ->
            Win

        _ ->
            Tie


goResult : Containers.User.User -> Result -> Int -> List Team -> Result
goResult u tw max ts =
    case ts of
        [] ->
            Lose

        t :: tts ->
            if u.id == t.playerOne || u.id == t.playerTwo then
                case compare t.score max of
                    EQ ->
                        tw

                    LT ->
                        Lose

                    _ ->
                        Lose

            else
                goResult u tw max tts


showTeams : List Team -> Html m
showTeams xs =
    div [] <| List.concatMap (\x -> [ text (showTeam x), br [] [] ]) xs


playerById : Int -> List User -> Maybe User
playerById n xs =
    case xs of
        [] ->
            Nothing

        x :: xxs ->
            if x.id == n then
                Just x

            else
                playerById n xxs


decodeTeam : Decoder Team
decodeTeam =
    Json.Decode.map3 Team
        (Json.Decode.field "FirstID" Json.Decode.int)
        (Json.Decode.field "SecondID" Json.Decode.int)
        (Json.Decode.field "Score" Json.Decode.int)


decode : Decoder Game
decode =
    Json.Decode.map7 Game
        (Json.Decode.field "ID" Json.Decode.int)
        (Json.Decode.field "NumPlayers" Json.Decode.int)
        (Json.Decode.field "Timer" Json.Decode.int)
        (Json.Decode.field "NumWords" Json.Decode.int)
        (Json.Decode.field "NumStages" Json.Decode.int)
        (Json.Decode.field "Host" Json.Decode.int)
        (Json.Decode.field "Players" (Json.Decode.list decodeUser))


show : Game -> String
show { id, numPlayers, timer, numWords, host, players } =
    String.join
        "\n"
        [ String.join " " [ "ID:", String.fromInt id ]
        , String.join " " [ "NumPlayers:", String.fromInt numPlayers ]
        , String.join " " [ "Timer:", String.fromInt timer ]
        , String.join " " [ "NumWords:", String.fromInt numWords ]
        , String.join " " [ "NumStages:", String.fromInt numWords ]
        , String.join " " [ "Host:", String.fromInt host ]
        , String.join " " <| "Players:" :: List.map Containers.User.show players
        ]


type alias LoginUser =
    { email : String, password : String }


encodeLogin : LoginUser -> Json.Encode.Value
encodeLogin { email, password } =
    Json.Encode.object
        [ ( "Email", Json.Encode.string email )
        , ( "Password", Json.Encode.string password )
        ]
