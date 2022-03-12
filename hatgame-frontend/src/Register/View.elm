module Register.View exposing (html)

import Html exposing (Html, br, button, div, input, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)
import UserCredentials


html : Maybe String -> UserCredentials.Data -> List (Html Msg)
html err registerData =
    [div [ class "container" ]
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
                        , ( "has-error", registerData.email == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Email" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , value (Maybe.withDefault "" registerData.email)
                        , onInput <|
                            \str ->
                                Msg.ChangeRegister
                                    { registerData
                                        | email = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    , div [ class "help-block" ]
                        [ text <| Html.Utils.printOnNothing registerData.email "Please enter an email"
                        ]
                    ]
                , div
                    [ classList
                        [ ( "form-group", True )
                        , ( "has-error", registerData.password == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Password" ]
                    , input
                        [ class "form-control"
                        , type_ "password"
                        , value (Maybe.withDefault "" registerData.password)
                        , onInput <|
                            \str ->
                                Msg.ChangeRegister
                                    { registerData
                                        | password = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    , div [ class "help-block" ]
                        [ text <| Html.Utils.printOnNothing registerData.password "Please enter a password"
                        ]
                    ]
                , div
                    [ class <|
                        "form-group"
                    ]
                    [ label [ class "control-label" ] [ text "Username" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , value (Maybe.withDefault "" registerData.username)
                        , onInput <|
                            \str ->
                                Msg.ChangeRegister
                                    { registerData
                                        | username = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    ]
                , button [ class "btn-primary", onClick Msg.Register ] [ text "Register" ]
                , Html.Utils.divOnJust err
                ]
            ]
        ]]
