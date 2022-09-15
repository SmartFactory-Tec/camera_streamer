#include "element.h"

bool setStatePlaying(GstElement* element) {
	return gst_element_set_state(element, GST_STATE_PLAYING) != GST_STATE_CHANGE_FAILURE;
}


bool setStatePaused(GstElement* element) {
    return gst_element_set_state(element, GST_STATE_PAUSED) != GST_STATE_CHANGE_FAILURE;
}

bool setStateReady(GstElement* element) {
    return gst_element_set_state(element, GST_STATE_READY) != GST_STATE_CHANGE_FAILURE;
}


bool setStateNull(GstElement* element) {
    return gst_element_set_state(element, GST_STATE_NULL) != GST_STATE_CHANGE_FAILURE;
}

char *getGstElementName(GstElement *element){
    return GST_ELEMENT_NAME(element);
}

void gst_set_string_property(GstElement *element, char *property_name, char *property_value) {
    g_object_set(element, property_name, property_value, NULL);
}

void gst_set_bool_property(GstElement *element, char *property_name, bool *property_value) {
    g_object_set(element, property_name, *property_value ? TRUE : FALSE, NULL);
}


