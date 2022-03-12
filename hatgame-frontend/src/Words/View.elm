module Words.View exposing (html)

import Containers.Game
import Html exposing (Html, button, div, h5, input, label, p, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)
import Words


html : Maybe String -> Words.Data -> List (Html Msg)
html err wordsData =
    [ div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    , ( "has-error", err /= Nothing )
                    ]
                ]
                [ div [ class "spacing-both" ] []
                , div
                    [ classList
                        [ ( "form-group", True )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Word" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , value (Maybe.withDefault "" wordsData.currentWord)
                        , onInput <|
                            \str ->
                                Msg.ChangeWords
                                    { wordsData
                                        | currentWord = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    , Html.Utils.divOnJust err
                    ]
                , button
                    [ class "btn-primary"
                    , onClick <| Msg.SendWord wordsData
                    , case wordsData.processState of
                        Words.Typing ->
                            disabled False

                        Words.Done ->
                            disabled True
                    ]
                    [ text "Send" ]
                , h5
                    []
                    [ text "Words" ]
                , div
                    [ class "display-window"
                    , style "height" <| String.fromInt (wordsData.game.numWords * 3) ++ "em"
                    , style "display" "flex"
                    , style
                        "flex-direction"
                        "column-reverse"
                    , style "justify-content" "space-around"
                    , style "align-items" "center"
                    ]
                  <|
                    List.map
                        (\str ->
                            label
                                [ class "user-label"
                                ]
                                [ text str ]
                        )
                        wordsData.words
                ]
            ]
        ]
    ]
