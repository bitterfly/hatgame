module Containers.Statistics exposing (Statistics, WordStats, decode, show)

import Html exposing (Html, div, dl, dt, h3, h4, h5, h6, label, p, table, tbody, td, text, th, thead, tr, ul)
import Html.Attributes exposing (..)
import Html.Utils
import Json.Decode exposing (Decoder)


type alias WordStats =
    { word : String
    , count : Int
    }


type alias Statistics =
    { gamesPlayed : Int
    , numberOfWins : Int
    , numberOfTies : Int
    , topWords : List WordStats
    }


decodeWords : Decoder WordStats
decodeWords =
    Json.Decode.map2 WordStats
        (Json.Decode.field "Word" Json.Decode.string)
        (Json.Decode.field "Count" Json.Decode.int)


decode : Decoder Statistics
decode =
    Json.Decode.map4 Statistics
        (Json.Decode.field "GamesPlayed" Json.Decode.int)
        (Json.Decode.field "NumberOfWins" Json.Decode.int)
        (Json.Decode.field "NumberOfTies" Json.Decode.int)
        (Json.Decode.field "TopWords" (Json.Decode.list decodeWords))


showWordStats : WordStats -> List (Html msg)
showWordStats { word, count } =
    [ tr [] [ th [ style "width" "90%" ] [ text word ], td [] [ text <| String.fromInt count ] ] ]



-- word ++ "[" ++ String.fromInt count ++ "]"


show : Maybe Statistics -> Html msg
show stat =
    case stat of
        Nothing ->
            text ""

        Just { gamesPlayed, numberOfWins, numberOfTies, topWords } ->
            let
                numberOfLosses =
                    gamesPlayed - numberOfWins - numberOfTies
            in
            div [ class "table-wrapper display-window" ] <|
                [ h6 [ style "textAlign" "left" ] [ text "Games" ]
                , p [] [ text <| String.fromInt gamesPlayed ]
                , table []
                    [ thead []
                        [ tr []
                            [ th []
                                [ text "Wins" ]
                            , th
                                []
                                [ text "Ties" ]
                            , th
                                []
                                [ text "Losses" ]
                            ]
                        ]
                    , tbody []
                        [ tr []
                            [ td []
                                [ text <| String.fromInt numberOfWins ]
                            , td
                                []
                                [ text <| String.fromInt numberOfTies ]
                            , td
                                []
                                [ text <| String.fromInt numberOfLosses ]
                            ]
                        ]
                    ]
                , Html.Utils.when
                    (numberOfLosses == 0 && gamesPlayed /= 0)
                    (h5 [ style "color" "red" ] [ text "LEGENDARY!" ])
                , h6
                    [ style "textAlign" "left" ]
                    [ text "Top words" ]
                , table []
                    [ div [ style "display" "grid", style "column-gap" "20px" ]
                        [ tbody [] <|
                            List.concat
                                (List.map
                                    showWordStats
                                    topWords
                                )
                        ]
                    ]
                ]
