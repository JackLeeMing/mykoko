FROM jumpserver/koko:v4.6.0-ce
RUN mv -f ./koko /opt/koko/koko_new
CMD [ "./koko_new" ]
