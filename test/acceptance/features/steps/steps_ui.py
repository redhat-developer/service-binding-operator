# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------

from behave import given, then, when, step


# STEP
@given(u'user [{username}] is logged into OpenShift Console with [{password}]')
def user_is_logged_into_openshift_console(username, password):
    raise NotImplementedError(f'STEP: Given user [{username}] is logged into OpenShift Console with [{password}]')


# STEP
@given(u'DB "{db_name}" is shown')
def db_is_shown(db_name):
    raise NotImplementedError(f'STEP: DB instance "{db_name}" is shown')


# STEP
@given(u'Application "{application_name}" is shown')
def app_is_shown(application_name):
    raise NotImplementedError(f'STEP: Application "{application_name}" is shown')


# STEP
arrow_is_dragged_from_application_to_db_step = u'Arrow is dragged and dropped from application icon to DB icon to "Create a binding connector"'


@given(arrow_is_dragged_from_application_to_db_step)
@when(arrow_is_dragged_from_application_to_db_step)
def arrow_is_dragged_from_application_to_db():
    raise NotImplementedError(u'STEP: Arrow is dragged and dropped from application icon to DB icon to "Create a binding connector"')


# STEP
arrow_is_rendered_from_application_to_db_step = u'Arrow is rendered from application icon to DB icon to indicate binding'


@step(arrow_is_rendered_from_application_to_db_step)
def arrow_is_rendered_from_application_to_db():
    raise NotImplementedError(u'STEP: Arrow is rendered from application icon to DB icon to indicate binding')


# STEP
@given(u'"{text}" is shown on page')
def text_is_shown_on_page(text):
    raise NotImplementedError(f'"{text}" is shown on page')


# STEP
@then(u'Operator page for "{operator_name}" is opened')
def operator_page_is_open(operator_name):
    raise NotImplementedError(f'STEP: Operator page for "{operator_name}" is opened')


# STEP
view_is_opened_step = u'"{view}" view is opened'


@given(view_is_opened_step)
@when(view_is_opened_step)
def view_is_opened(view):
    raise NotImplementedError(f'STEP: "{view}" view is opened')


# STEP
page_is_opened_step = u'"{page}" page is opened'


@given(page_is_opened_step)
@when(page_is_opened_step)
def page_is_opened(page):
    raise NotImplementedError(f'STEP: "{page}" page is opened')


# STEP
@given(u'"{card}" card is clicked and Community operators confirmed')
def card_is_clicked(card):
    raise NotImplementedError(f'STEP: Given "{card}" card is clicked and Community operators confirmed')


# STEP
@given(u'"{selection}" checkbox for "{selection_category}" is selected')
def step_impl(selection, selection_category):
    raise NotImplementedError(f'STEP: Given "{selection}" checkbox for "{selection_category}" is selected')


# STEP
button_is_clicked_step = u'"{button}" button is clicked'


@given(button_is_clicked_step)
@when(button_is_clicked_step)
def button_is_clicked(button):
    raise NotImplementedError(f'STEP: "{button}" button is clicked')


# STEP
selection_is_selected_step = '"{selection}" is selected to be "{selected_value}"'


@given(selection_is_selected_step)
@when(selection_is_selected_step)
def selection_is_selected(selection, selected_value):
    raise NotImplementedError(f'STEP: "{selection}" is selected to be "{selected_value}"')


# STEP
@then(u'Operator status is "{status}"')
def operator_status_is(status):
    raise NotImplementedError(f'STEP: Operator status is "{status}"')
