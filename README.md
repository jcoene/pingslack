# Pingslack

Relays notifications from Pingdom to Slack. Suitable for deployment on Heroku.

## Deploying on Heroku:

Clone this repository:

    git clone git@github.com:jcoene/pingslack.git && cd pingslack

Make sure you're logged into Heroku:

    heroku login

You'll need to specify the Heroku Go buildpack during app creation:

    heroku create mycompany-pingslack -b https://github.com/kr/heroku-buildpack-go.git

Make sure you set the SLACK_TOKEN, SLACK_CHANNEL and SLACK_DOMAIN environment variables:

    heroku set SLACK_TOKEN=my_token SLACK_CHANNEL=operations SLACK_DOMAIN=my_domain.slack.com

Then deploy:

    git push heroku master

## Configuring Pingdom

1. Sign into Pingdom. Navigate to **Alerting** -> **BeepManager** -> **Users**.
2. **Add** or **Edit** a user that has the alerting settings you desire for the notifications.
3. Under **Contact Methods**, add a new **Webhook/URL** and use `http://mycompany-pingslack.herokuapp.com/notify` for the Webhook URL (replace *mycompany-pingslack* with the name of your heroku app). Save the new Webhook contact method.
4. Choose **Old Message Format** for the Alert Message format setting and save the user.

## License

Copyright (C) 2014 Jason Coene. MIT License. See LICENSE for details.
