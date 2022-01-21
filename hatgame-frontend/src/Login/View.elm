module Login.View exposing (html)

import Html exposing (Html, button, div, img, input, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Login
import Maybe.Utils
import Msg exposing (Msg)


html : Maybe String -> Login.Data -> Html Msg
html err loginData =
    div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    , ( "has-error", err /= Nothing )
                    ]
                ]
                [ div
                    [ class "header-gif" ]
                    [ img [ src "hatgame.gif", style "height" "25em" ] [] ]
                , div [ class "spacing-both" ] []
                , div
                    [ classList
                        [ ( "form-group", True )
                        , ( "has-error", loginData.email == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Email" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , value (Maybe.withDefault "" loginData.email)
                        , onInput <|
                            \str ->
                                Msg.ChangeLogin
                                    { loginData
                                        | email = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    , div [ class "help-block" ]
                        [ text <| Html.Utils.printOnNothing loginData.email "Please enter your email"
                        ]
                    ]
                , div
                    [ classList
                        [ ( "form-group", True )
                        , ( "has-error", loginData.password == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Password" ]
                    , input
                        [ class "form-control"
                        , type_ "password"
                        , value (Maybe.withDefault "" loginData.password)
                        , onInput <|
                            \str ->
                                Msg.ChangeLogin
                                    { loginData
                                        | password = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    , div [ class "help-block" ]
                        [ text <| Html.Utils.printOnNothing loginData.password "Please enter your password"
                        ]
                    ]
                , button
                    [ class "btn-primary"
                    , onClick Msg.Login
                    , disabled <| loginData.email == Nothing || loginData.password == Nothing
                    ]
                    [ text "Login" ]
                , button
                    [ class "btn-secondary"
                    , onClick Msg.ToRegisterPage
                    ]
                    [ text "Register" ]
                , Html.Utils.divOnJust err
                ]
            ]
        ]



--     , br [] []
--     ]
