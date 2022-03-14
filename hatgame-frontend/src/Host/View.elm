module Host.View exposing (html)

import Host
import Html exposing (Html, button, div, input, label, p, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)
import Page exposing (Page)


html : Maybe String -> Host.Data -> List (Html Msg)
html err hostData =
    [ div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    , ( "has-error", err /= Nothing )
                    ]
                ]
                [ div [ class "form-group" ]
                    [ div [ class "spacing-both" ] []
                    , label [ class "control-label" ] [ text "Players" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , style "text-align" "center"
                        , readonly True
                        , value <| String.fromInt hostData.players
                        ]
                        []
                    , div
                        []
                        [ button
                            [ class "btn-secondary"
                            , onClick <| Msg.ChangeHost { hostData | players = hostData.players + 2 }
                            ]
                            [ p
                                [ class "arrow up"
                                ]
                                []
                            ]
                        , button
                            [ class "btn-secondary"
                            , onClick <|
                                if hostData.players == Host.default.players then
                                    Msg.Nothing

                                else
                                    Msg.ChangeHost { hostData | players = hostData.players - 2 }
                            ]
                            [ p
                                [ class "arrow down"
                                ]
                                []
                            ]
                        ]
                    , div [ class "spacing-both" ] []
                    , label [ class "control-label" ] [ text "Words" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , style "text-align" "center"
                        , value <| Maybe.withDefault "" (Maybe.map String.fromInt hostData.words)
                        , onInput <|
                            \str ->
                                Msg.ChangeHost <|
                                    { hostData
                                        | words =
                                            case Maybe.Utils.toMaybeInt str of
                                                Maybe.Utils.NoString ->
                                                    Nothing

                                                Maybe.Utils.NotInt ->
                                                    hostData.words

                                                Maybe.Utils.ParsedInt n ->
                                                    Just n
                                    }
                        ]
                        []
                    , div [ class "spacing-both" ] []
                    , label [ class "control-label" ] [ text "Stages" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , style "text-align" "center"
                        , readonly True
                        , value <| String.fromInt hostData.stages
                        ]
                        []
                    , div
                        []
                        [ button
                            [ class "btn-secondary"
                            , onClick <|
                                if hostData.stages == Host.default.maxStages then
                                    Msg.Nothing

                                else
                                    Msg.ChangeHost { hostData | stages = hostData.stages + 1 }
                            ]
                            [ p
                                [ class "arrow up"
                                ]
                                []
                            ]
                        , button
                            [ class "btn-secondary"
                            , onClick <|
                                if hostData.stages == Host.default.stages then
                                    Msg.Nothing

                                else
                                    Msg.ChangeHost { hostData | stages = hostData.stages - 1 }
                            ]
                            [ p
                                [ class "arrow down"
                                ]
                                []
                            ]
                        ]
                    , div [ class "spacing-both" ] []
                    , label [ class "control-label" ] [ text "Seconds" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , style "text-align" "center"
                        , value <| Maybe.withDefault "" (Maybe.map String.fromInt hostData.timer)
                        , onInput <|
                            \str ->
                                Msg.ChangeHost <|
                                    { hostData
                                        | timer =
                                            case Maybe.Utils.toMaybeInt str of
                                                Maybe.Utils.NoString ->
                                                    Nothing

                                                Maybe.Utils.NotInt ->
                                                    hostData.timer

                                                Maybe.Utils.ParsedInt n ->
                                                    Just n
                                    }
                        ]
                        []
                    , button
                        [ class "btn-primary"
                        , onClick Msg.Host
                        ]
                        [ text "Start" ]
                    , Html.Utils.divOnJust err
                    , button
                        [ class "btn-primary"
                        , onClick <|
                            Msg.GoTo
                                (Page.Home
                                    { gameId = Nothing
                                    , stats = Nothing
                                    }
                                )
                        ]
                        [ text "Quit" ]
                    ]
                ]
            ]
        ]
    ]
