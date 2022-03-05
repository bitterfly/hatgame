module Started.View exposing (html)

import Containers.Game
import Containers.User
import Generic.Utils
import Html exposing (Html, button, div, h1, h5, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Msg exposing (Msg)
import Progress.Ring
import Started


html : Maybe Containers.User.WithToken -> Maybe String -> Started.Data -> Html Msg
html tokenUser err startedData =
    div []
        [ div [ class "header-home" ]
            [ label [ class "user-label" ] [ text <| Containers.User.maybeUsernameWithToken tokenUser ]
            , text "you're playing with"
            , label
                [ class "user-label" ]
                [ text <| Containers.User.maybeUsername startedData.partner ]
            ]
        , div [ class "container" ]
            [ div [ class "row" ]
                [ div
                    [ classList
                        [ ( "shift-3", True )
                        , ( "has-error", err /= Nothing )
                        ]
                    ]
                    [ div
                        [ style "display" "flex"
                        , style
                            "flex-direction"
                            "column"
                        , style "justify-content" "space-around"
                        , style "align-items" "center"
                        ]
                        [ div [ class "spacing-both" ] []
                        , Progress.Ring.view
                            { color =
                                Generic.Utils.makeColor
                                    (toFloat (Maybe.withDefault startedData.game.timer startedData.timer)
                                        / toFloat startedData.game.timer
                                    )
                            , progress =
                                toFloat
                                    (Maybe.withDefault startedData.game.timer startedData.timer)
                                    / toFloat startedData.game.timer
                            , stroke = 30
                            , radius = 100
                            }
                        , div [ class "spacing-both" ] []
                        , case startedData.processState of
                            Started.NotStoryteller user ->
                                div []
                                    [ h5 []
                                        [ text <|
                                            case user of
                                                Nothing ->
                                                    "Watiing for storyteller."

                                                Just u ->
                                                    case startedData.partner of
                                                        Nothing ->
                                                            ""

                                                        Just p ->
                                                            if u.id == p.id then
                                                                Containers.User.show u ++ "'s turn. You're guessing."

                                                            else
                                                                Containers.User.show u ++ "'s turn. Shhh, just listen!"
                                        ]
                                    ]

                            Started.StorytellerWaiting ->
                                div
                                    [ style "display" "flex"
                                    , style "flex-direction" "column"
                                    , style "justify-content" "center"
                                    , style "align-items" "center"
                                    ]
                                    [ button
                                        [ class "btn-primary"
                                        , onClick <| Msg.SendReadyStoryteller
                                        ]
                                        [ text "Start" ]
                                    ]

                            Started.StorytellerActive ->
                                div
                                    [ style "display" "flex"
                                    , style "flex-direction" "column"
                                    , style "justify-content" "center"
                                    , style "align-items" "center"
                                    ]
                                    [ h1 [] [ text <| Maybe.withDefault "" startedData.currentWord ]
                                    , button
                                        [ class "btn-primary"
                                        , case startedData.currentWord of
                                            Nothing ->
                                                disabled True

                                            Just word ->
                                                onClick <| Msg.SendGuessed word
                                        ]
                                        [ text "Yep!" ]
                                    ]
                        ]
                    ]
                ]
            ]
        ]
