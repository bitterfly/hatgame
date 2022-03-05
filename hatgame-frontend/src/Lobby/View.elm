module Lobby.View exposing (html)

import Html exposing (Html, button, div, h5, label, p, table, tbody, td, text, th, tr)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Lobby
import Msg exposing (Msg)


html : Lobby.Data -> Html Msg
html { game, processState } =
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
                                [ text <| String.fromInt game.id ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Timer" ]
                            , p []
                                [ text <|
                                    String.fromInt game.timer
                                ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Words" ]
                            , p []
                                [ text <|
                                    String.fromInt game.numWords
                                ]
                            ]
                        , div
                            [ style "display" "flex"
                            , style "justify-content" "space-around"
                            ]
                            [ p [] [ text "Players" ]
                            , p []
                                [ text <|
                                    String.fromInt (List.length game.players)
                                        ++ " / "
                                        ++ String.fromInt game.numPlayers
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
                                        ++ (String.join ", " <| List.map (\u -> u.username) game.players)
                                        ++ " ]"
                                ]
                            ]
                        ]
                    ]
                , case processState of
                    Lobby.WaitingPlayers ->
                        div
                            []
                            []

                    Lobby.ReadyToStart ->
                        div
                            [ style "display" "flex"
                            , style "flex-direction" "column"
                            , style "justify-content" "center"
                            , style "align-items" "center"
                            ]
                            [ button
                                [ class "btn-primary"
                                , onClick <| Msg.SendRequestToStart
                                ]
                                [ text "Start" ]
                            ]
                ]
            ]
        ]
