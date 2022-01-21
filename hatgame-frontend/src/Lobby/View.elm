module Lobby.View exposing (html)

import Containers.Game
import Host
import Html exposing (Html, button, div, h5, label, p, table, tbody, td, text, th, tr)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)


html : Containers.Game.Game -> Html msg
html { id, numPlayers, timer, numWords, players } =
    div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ class "shift-3"
                ]
                [ div [ class "spacing-both" ]
                    []
                , div
                    [ style "display" "flex"
                    , style "justify-content" "center"
                    ]
                    [ h5
                        []
                        [ text "Waiting for the other players" ]
                    ]
                , div [ class "spacing-both" ] []
                , div [ class "display-window" ]
                    [ div
                        []
                        [ div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p []
                                [ text "Room" ]
                            , p []
                                [ text <| String.fromInt id ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Timer" ]
                            , p []
                                [ text <|
                                    String.fromInt timer
                                ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Words" ]
                            , p []
                                [ text <|
                                    String.fromInt numWords
                                ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Players" ]
                            , p []
                                [ text <|
                                    String.fromInt (List.length players)
                                        ++ " / "
                                        ++ String.fromInt numPlayers
                                ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "center"
                            ]
                            [ label
                                [ class "user-label"
                                ]
                                [ text <|
                                    "[ "
                                        ++ (String.join ", " <| List.map (\u -> u.username) players)
                                        ++ " ]"
                                ]
                            ]
                        ]
                    ]
                ]
            ]
        ]
