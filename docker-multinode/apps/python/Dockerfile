FROM alpine:3.5

RUN apk add --update musl linux-headers build-base dumb-init protobuf gcc protobuf-dev python2 bash git ca-certificates python2-dev py-setuptools tzdata
RUN cp /usr/share/zoneinfo/Europe/Brussels /etc/localtime
RUN echo "Asia/Hong_Kong" >  /etc/timezone
RUN echo \

  && if [[ ! -e /usr/bin/python ]];        then ln -sf /usr/bin/python2.7 /usr/bin/python; fi \
  && if [[ ! -e /usr/bin/python-config ]]; then ln -sf /usr/bin/python2.7-config /usr/bin/python-config; fi \
  && if [[ ! -e /usr/bin/easy_install ]];  then ln -sf /usr/bin/easy_install-2.7 /usr/bin/easy_install; fi \

  # Install and upgrade Pip
  && easy_install pip \
  && pip install --upgrade pip \
  && if [[ ! -e /usr/bin/pip ]]; then ln -sf /usr/bin/pip2.7 /usr/bin/pip; fi \
  && echo
RUN easy_install pip
RUN pip install --upgrade pip
RUN pip install twisted httplib2 protobuf
CMD ["sh"]

