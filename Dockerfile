FROM ruby:2.2.0
COPY ./vquery/Gemfile /tmp/Gemfile
WORKDIR /tmp
RUN bundle install
COPY ./vquery /vquery
COPY /vqserver /vqserver
WORKDIR /
ENTRYPOINT ["/vqserver"]
