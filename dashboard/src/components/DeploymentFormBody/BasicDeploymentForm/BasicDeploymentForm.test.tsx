import * as React from "react";

import { mount, shallow } from "enzyme";
import { DeploymentEvent, IBasicFormParam, IBasicFormSliderParam } from "shared/types";
import BasicDeploymentForm from "./BasicDeploymentForm";
import Subsection from "./Subsection";

const defaultProps = {
  deploymentEvent: "install" as DeploymentEvent,
  params: [],
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  appValues: "",
  handleValuesChange: jest.fn(),
};

[
  {
    description: "renders a basic deployment with a username",
    params: [{ path: "wordpressUsername", value: "user" } as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a password",
    params: [{ path: "wordpressPassword", value: "sserpdrow" } as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a email",
    params: [{ path: "wordpressEmail", value: "user@example.com" } as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a generic string",
    params: [{ path: "blogName", value: "my-blog", type: "string" } as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with custom configuration",
    params: [
      {
        path: "configuration",
        value: "First line\nSecond line",
        render: "textArea",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a disk size",
    params: [
      {
        path: "size",
        value: "10Gi",
        type: "string",
        render: "slider",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a integer disk size",
    params: [
      {
        path: "size",
        value: 10,
        type: "integer",
        render: "slider",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a number disk size",
    params: [
      {
        path: "size",
        value: 10.0,
        type: "number",
        render: "slider",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with slider parameters",
    params: [
      {
        path: "size",
        value: "10Gi",
        type: "string",
        render: "slider",
        sliderMin: 1,
        sliderMax: 100,
        sliderStep: 1,
        sliderUnit: "Gi",
      } as IBasicFormSliderParam,
    ],
  },
  {
    description: "renders a basic deployment with username, password, email and a generic string",
    params: [
      { path: "wordpressUsername", value: "user" } as IBasicFormParam,
      { path: "wordpressPassword", value: "sserpdrow" } as IBasicFormParam,
      { path: "wordpressEmail", value: "user@example.com" } as IBasicFormParam,
      { path: "blogName", value: "my-blog", type: "string" } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a generic boolean",
    params: [{ path: "enableMetrics", value: true, type: "boolean" } as IBasicFormParam],
  },
  {
    description: "renders a basic deployment with a generic number",
    params: [{ path: "replicas", value: 1, type: "integer" } as IBasicFormParam],
  },
].forEach(t => {
  it(t.description, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <BasicDeploymentForm
        {...defaultProps}
        params={t.params}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );
    expect(wrapper).toMatchSnapshot();

    t.params.forEach((param, i) => {
      wrapper
        .find(`${param.render === "textArea" ? "textarea" : "input"}#${param.path}-${i}`)
        .simulate("change");
      const mockCalls = handleBasicFormParamChange.mock.calls;
      expect(mockCalls[i]).toEqual([param]);
      expect(onChange.mock.calls.length).toBe(i + 1);
    });
  });
});

it("should render an external database section", () => {
  const params = [
    {
      path: "edbs",
      value: {},
      type: "object",
      children: [{ path: "mariadb.enabled", value: {}, type: "boolean" }],
    } as IBasicFormParam,
  ];
  const wrapper = mount(<BasicDeploymentForm {...defaultProps} params={params} />);

  const dbsec = wrapper.find(Subsection);
  expect(dbsec).toExist();
});

it("should hide an element if it depends on a param (string)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: "bar",
    },
    {
      path: "bar",
      type: "boolean",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: true";
  const wrapper = shallow(
    <BasicDeploymentForm {...defaultProps} params={params} appValues={appValues} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should hide an element if it depends on a single param (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        value: "enabled",
        path: "bar",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: enabled";
  const wrapper = shallow(
    <BasicDeploymentForm {...defaultProps} params={params} appValues={appValues} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should hide an element if it depends on multiple params (AND) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "and",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: enabled\nbaz: disabled";
  const wrapper = shallow(
    <BasicDeploymentForm {...defaultProps} params={params} appValues={appValues} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should hide an element if it depends on multiple params (OR) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "or",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: enabled\nbaz: enabled";
  const wrapper = shallow(
    <BasicDeploymentForm {...defaultProps} params={params} appValues={appValues} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should hide an element if it depends on multiple params (NOR) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            value: "enabled",
            path: "bar",
          },
          {
            value: "disabled",
            path: "baz",
          },
        ],
        operator: "nor",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: disabled\nbaz: enabled";
  const wrapper = shallow(
    <BasicDeploymentForm {...defaultProps} params={params} appValues={appValues} />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should hide an element if it depends on the deploymentEvent (install | upgrade) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        event: "upgrade",
      },
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: disabled\nbaz: enabled";
  const wrapper = shallow(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="upgrade"
      params={params}
      appValues={appValues}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});

it("should NOT hide an element if it depends on the deploymentEvent (install | upgrade) (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        event: "upgrade",
      },
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: disabled\nbaz: enabled";
  const wrapper = shallow(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="install"
      params={params}
      appValues={appValues}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).not.toExist();
});

it("should hide an element if it depends on deploymentEvent (install | upgrade) combined with multiple params (object)", () => {
  const params = [
    {
      path: "foo",
      type: "string",
      hidden: {
        conditions: [
          {
            event: "upgrade",
          },
          {
            value: "enabled",
            path: "bar",
          },
        ],
        operator: "or",
      },
    },
    {
      path: "bar",
      type: "string",
    },
  ] as IBasicFormParam[];
  const appValues = "foo: 1\nbar: disabled";
  const wrapper = shallow(
    <BasicDeploymentForm
      {...defaultProps}
      deploymentEvent="upgrade"
      params={params}
      appValues={appValues}
    />,
  );

  const hiddenParam = wrapper.find("div").filterWhere(p => p.prop("hidden") === true);
  expect(hiddenParam).toExist();
});
